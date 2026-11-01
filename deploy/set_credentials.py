#!/usr/bin/env python3
"""Rotate the site owner's login. Read credentials from stdin, never argv."""

import hashlib
import json
import os
import re
import secrets
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from urllib.parse import urlsplit
from datetime import datetime, timezone
from pathlib import Path


def atomic_write(path, content, mode):
    descriptor, temporary = tempfile.mkstemp(prefix=".updating-", dir=path.parent)
    try:
        with os.fdopen(descriptor, "w", encoding="utf-8") as stream:
            stream.write(content)
            stream.flush()
            os.fsync(stream.fileno())
        info = path.stat() if path.exists() else None
        os.chmod(temporary, mode)
        if info:
            os.chown(temporary, info.st_uid, info.st_gid)
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary):
            os.unlink(temporary)


def main():
    if os.geteuid() != 0:
        raise RuntimeError("Run as root")
    print("Ready for credential JSON on stdin (not echoed).", flush=True)
    payload = json.loads(sys.stdin.readline())
    username, password = payload.get("username"), payload.get("password")
    if not isinstance(username, str) or not re.fullmatch(r"[A-Za-z0-9_.-]{1,64}", username):
        raise RuntimeError("Invalid username")
    if not isinstance(password, str) or not 8 <= len(password) <= 1024:
        raise RuntimeError("Invalid password length")
    config = Path("/etc/devgraph")
    access_path, runtime_path = config / "access.json", config / "runtime.env"
    previous_access = json.loads(access_path.read_text())
    runtime = dict(line.split("=", 1) for line in runtime_path.read_text().splitlines() if "=" in line and not line.startswith("#"))
    origin = runtime["DEVGRAPH_PUBLIC_ORIGIN"]
    host = urlsplit(origin).netloc

    def request(path, data=None, cookie=None):
        headers = {"Host": host, "Origin": origin, "X-DevGraph-Request": "1", "Content-Type": "application/json"}
        if cookie:
            headers["Cookie"] = cookie
        req = urllib.request.Request("http://127.0.0.1:8765" + path, data=json.dumps(data).encode() if data is not None else None, headers=headers)
        try:
            with urllib.request.urlopen(req, timeout=10) as response:
                return response.status, response.headers, response.read()
        except urllib.error.HTTPError as error:
            return error.code, error.headers, error.read()

    status, headers, _ = request("/api/login", {"username": previous_access["username"], "password": previous_access["password"]})
    old_cookie = headers.get("Set-Cookie", "").split(";", 1)[0] if status == 200 else ""
    backup = Path("/opt/backups/devgraph") / ("credentials-" + datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ"))
    backup.mkdir(parents=True, mode=0o700)
    os.chmod(backup, 0o700)
    files = [access_path, runtime_path]
    htpasswd = config / "htpasswd"
    if htpasswd.exists():
        files.append(htpasswd)
    for path in files:
        shutil.copy2(path, backup / path.name)

    salt = secrets.token_hex(16)
    runtime.update({"DEVGRAPH_AUTH_ENABLED": "1", "DEVGRAPH_USERNAME": username,
                    "DEVGRAPH_PASSWORD_SALT": salt,
                    "DEVGRAPH_PASSWORD_HASH": hashlib.pbkdf2_hmac("sha256", password.encode(), bytes.fromhex(salt), 260000).hex()})
    access = {**previous_access, "username": username, "password": password}
    try:
        atomic_write(runtime_path, "\n".join(f"{key}={value}" for key, value in runtime.items()) + "\n", 0o640)
        atomic_write(access_path, json.dumps(access, indent=2) + "\n", 0o600)
        if htpasswd.exists():
            hashed = subprocess.run(["openssl", "passwd", "-6", "-stdin"], input=password + "\n", text=True, capture_output=True, check=True).stdout.strip()
            atomic_write(htpasswd, username + ":" + hashed + "\n", 0o640)
        subprocess.run(["systemctl", "restart", "devgraph"], check=True)
        for attempt in range(30):
            try:
                status, headers, _ = request("/api/login", {"username": username, "password": password})
                break
            except OSError:
                if attempt == 29:
                    raise
                time.sleep(.2)
        if status != 200:
            raise RuntimeError("Updated login failed verification")
        cookie = headers.get("Set-Cookie", "").split(";", 1)[0]
        status, _, body = request("/api/notes", cookie=cookie)
        if status != 200:
            raise RuntimeError("Authenticated read failed verification")
        count = len(json.loads(body)["notes"])
        if old_cookie and request("/api/notes", cookie=old_cookie)[0] != 401:
            raise RuntimeError("Old session was not revoked")
        if (username, password) != (previous_access["username"], previous_access["password"]):
            if request("/api/login", {"username": previous_access["username"], "password": previous_access["password"]})[0] != 401:
                raise RuntimeError("Old credentials were not revoked")
    except Exception:
        for path in files:
            shutil.copy2(backup / path.name, path)
        subprocess.run(["systemctl", "restart", "devgraph"], check=True)
        raise
    print(json.dumps({"username": username, "login_verified": True, "old_sessions_revoked": True,
                      "notes": count, "backup": str(backup)}), flush=True)


if __name__ == "__main__":
    main()
