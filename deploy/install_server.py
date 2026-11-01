#!/usr/bin/env python3
"""Install a release behind an existing Nginx vhost without removing other apps."""

import argparse
import grp
import hashlib
import json
import os
import pwd
import secrets
import shutil
import subprocess
import tarfile
from datetime import datetime, timezone
from pathlib import Path


def run(*command, **kwargs):
    return subprocess.run(command, check=True, text=True, **kwargs)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--archive", required=True, type=Path)
    parser.add_argument("--sha256", required=True)
    parser.add_argument("--site-config", required=True, type=Path)
    parser.add_argument("--origin", required=True)
    parser.add_argument("--mount", default="/graph/")
    args = parser.parse_args()
    if os.geteuid() != 0:
        parser.error("Run as root.")
    if hashlib.sha256(args.archive.read_bytes()).hexdigest() != args.sha256:
        parser.error("Archive checksum does not match.")
    if args.mount != "/graph/" or args.origin != "https://101.34.211.31":
        parser.error("This deployment is scoped to https://101.34.211.31/graph/.")

    release_id = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    release = Path("/opt/devgraph/releases") / release_id
    backup = Path("/opt/backups/devgraph") / release_id
    backup.mkdir(parents=True, mode=0o700)
    os.chmod(backup, 0o700)
    original = args.site_config.read_text()
    shutil.copy2(args.site_config, backup / args.site_config.name)
    for name in ["devgraph.service"]:
        path = Path("/etc/systemd/system") / name
        if path.exists():
            shutil.copy2(path, backup / name)
    old_runtime = Path("/etc/devgraph/runtime.env")
    if old_runtime.exists():
        shutil.copy2(old_runtime, backup / "runtime.env")
    previous = Path("/opt/devgraph/current")
    previous_target = str(previous.resolve()) if previous.exists() else None
    (backup / "state.json").write_text(json.dumps({"previous_release": previous_target, "site_config": str(args.site_config)}, indent=2))

    release.mkdir(parents=True)
    with tarfile.open(args.archive) as archive:
        for member in archive.getmembers():
            target = (release / member.name).resolve()
            if not target.is_relative_to(release) or not (member.isfile() or member.isdir()):
                raise RuntimeError("Unsafe archive member")
        archive.extractall(release)
    run("/usr/bin/python3", "-m", "unittest", "discover", "-s", "tests", "-v", cwd=release)
    try:
        user = pwd.getpwnam("devgraph")
    except KeyError:
        run("useradd", "--system", "--home-dir", "/var/lib/devgraph", "--shell", "/sbin/nologin", "devgraph")
        user = pwd.getpwnam("devgraph")
    state = Path("/var/lib/devgraph")
    state.mkdir(exist_ok=True, mode=0o750)
    vault = state / "vault"
    if not vault.exists():
        shutil.copytree(release / "seed-vault", vault)
    for path in [state, *state.rglob("*")]:
        os.chown(path, user.pw_uid, user.pw_gid)
        os.chmod(path, 0o750 if path.is_dir() else 0o640)

    config = Path("/etc/devgraph")
    config.mkdir(exist_ok=True, mode=0o750)
    nginx_gid = grp.getgrnam("nginx").gr_gid
    os.chown(config, 0, nginx_gid)
    access_file = config / "access.json"
    if not access_file.exists():
        password = secrets.token_urlsafe(24)
        access_file.write_text(json.dumps({"url": args.origin + args.mount, "username": "devgraph", "password": password}, indent=2))
        os.chmod(access_file, 0o600)
    access = json.loads(access_file.read_text())
    password_hash = run("openssl", "passwd", "-6", "-stdin", input=access["password"] + "\n", capture_output=True).stdout.strip()
    password_file = config / "htpasswd"
    password_file.write_text(access["username"] + ":" + password_hash + "\n")
    os.chmod(password_file, 0o640)
    os.chown(password_file, 0, nginx_gid)
    runtime = config / "runtime.env"
    salt = secrets.token_hex(16)
    digest = hashlib.pbkdf2_hmac("sha256", access["password"].encode(), bytes.fromhex(salt), 260000).hex()
    runtime.write_text("\n".join(["DEVGRAPH_PUBLIC_ORIGIN=" + args.origin,
                                "DEVGRAPH_AUTH_ENABLED=1", "DEVGRAPH_USERNAME=" + access["username"],
                                "DEVGRAPH_PASSWORD_HASH=" + digest, "DEVGRAPH_PASSWORD_SALT=" + salt,
                                "DEVGRAPH_COOKIE_PATH=" + args.mount, ""]))
    os.chmod(runtime, 0o640)

    next_link = Path("/opt/devgraph/.current-next")
    if next_link.is_symlink():
        next_link.unlink()
    next_link.symlink_to(release)
    os.replace(next_link, previous)
    unit = Path("/etc/systemd/system/devgraph.service")
    shutil.copy2(release / "deploy/devgraph.service.example", unit)
    run("systemctl", "daemon-reload")
    run("systemctl", "enable", "--now", "devgraph.service")
    run("systemctl", "restart", "devgraph.service")

    location = '''
    # DevGraph: separate authentication and persistent Markdown storage.
    location = /graph { return 308 /graph/; }
    location ^~ /graph/ {
        auth_basic off;
        client_max_body_size 2m;
        proxy_pass http://127.0.0.1:8765/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 5s;
        proxy_read_timeout 30s;
    }
'''
    changed = original
    if "location ^~ /graph/" not in original:
        index = original.find("    location / {", original.find("listen 443"))
        if index < 0:
            raise RuntimeError("Expected HTTPS location was not found; original Nginx configuration is unchanged.")
        changed = original[:index] + location + original[index:]
    else:
        changed = original.replace('auth_basic "DevGraph";\n        auth_basic_user_file /etc/devgraph/htpasswd;', 'auth_basic off;')
    args.site_config.write_text(changed)
    try:
        run("nginx", "-t")
        run("systemctl", "reload", "nginx")
    except Exception:
        args.site_config.write_text(original)
        run("nginx", "-t")
        run("systemctl", "reload", "nginx")
        raise
    print(json.dumps({"release": str(release), "backup": str(backup), "vault": str(vault), "access_file": str(access_file), "url": access["url"]}))


if __name__ == "__main__":
    main()
