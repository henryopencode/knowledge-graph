#!/usr/bin/env python3
"""Back up and retire only the two projects explicitly selected by the owner."""

import hashlib
import json
import os
import shutil
import signal
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path

PROJECTS = ("ebmscreen", "litdemo")


def manifest(root):
    result = {}
    for path in sorted(root.rglob("*")):
        name = path.relative_to(root).as_posix()
        if path.is_symlink():
            result[name] = {"link": os.readlink(path)}
        elif path.is_file():
            digest = hashlib.sha256()
            with path.open("rb") as stream:
                for chunk in iter(lambda: stream.read(1024 * 1024), b""):
                    digest.update(chunk)
            result[name] = {"bytes": path.stat().st_size, "sha256": digest.hexdigest()}
    return result


def main():
    if os.geteuid() != 0:
        raise RuntimeError("Run as root")
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    backup = Path("/opt/backups/devgraph") / ("retired-projects-" + stamp)
    backup.mkdir(parents=True, mode=0o700)
    os.chmod(backup, 0o700)
    state = {}
    for name in PROJECTS:
        source = Path("/opt") / name
        if not source.is_dir() or source.is_symlink():
            raise RuntimeError(f"Unexpected project path: {source}")
        instruction = source / "AGENTS.md"
        if instruction.exists():
            raise RuntimeError(f"Additional project instructions require inspection: {instruction}")
        state[name] = {"enabled": subprocess.run(["systemctl", "is-enabled", name], capture_output=True, text=True).stdout.strip(),
                       "active": subprocess.run(["systemctl", "is-active", name], capture_output=True, text=True).stdout.strip()}
        for source_file in [Path("/etc/systemd/system") / (name + ".service"), Path("/etc/nginx/conf.d") / (name + ".conf")]:
            if source_file.exists():
                destination = backup / "config" / source_file.relative_to("/")
                destination.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(source_file, destination)
                if source_file.suffix == ".conf":
                    for line in source_file.read_text().splitlines():
                        if line.strip().startswith("auth_basic_user_file "):
                            password_file = Path(line.strip().split()[1].rstrip(";"))
                            if password_file.is_file():
                                saved = backup / "config" / password_file.relative_to("/")
                                saved.parent.mkdir(parents=True, exist_ok=True)
                                shutil.copy2(password_file, saved)
    (backup / "service-state.json").write_text(json.dumps(state, indent=2))
    subprocess.run(["systemctl", "disable", "--now", *PROJECTS], check=True)
    # Old development servers may be outside the systemd service cgroup.
    for proc in Path("/proc").iterdir():
        if not proc.name.isdigit():
            continue
        try:
            cwd = Path(os.readlink(proc / "cwd"))
            if any(cwd == Path("/opt") / name or cwd.is_relative_to(Path("/opt") / name) for name in PROJECTS):
                pid = int(proc.name)
                os.kill(pid, signal.SIGTERM)
                for _ in range(100):
                    if not proc.exists():
                        break
                    time.sleep(.1)
                if proc.exists() and Path(os.readlink(proc / "cwd")) == cwd:
                    os.kill(pid, signal.SIGKILL)
                print(f"Stopped retired-project process {pid}", flush=True)
        except (FileNotFoundError, ProcessLookupError, PermissionError):
            continue
    try:
        for name in PROJECTS:
            source = Path("/opt") / name
            destination = backup / "projects" / name
            destination.parent.mkdir(exist_ok=True)
            shutil.copytree(source, destination, symlinks=True)
            expected = manifest(source)
            actual = manifest(destination)
            if expected != actual:
                raise RuntimeError(f"Backup verification failed for {name}; source remains intact")
            (backup / (name + "-manifest.json")).write_text(json.dumps(expected, ensure_ascii=False, indent=2))
            print(f"Verified backup for {name}: {len(expected)} files/links", flush=True)
    except Exception:
        for name, previous in state.items():
            if previous["enabled"] == "enabled":
                subprocess.run(["systemctl", "enable", name], check=True)
            if previous["active"] == "active":
                subprocess.run(["systemctl", "start", name], check=True)
        raise
    # All original project files are removed only after both backups match.
    for name in PROJECTS:
        for path in [Path("/etc/nginx/conf.d") / (name + ".conf"), Path("/etc/systemd/system") / (name + ".service")]:
            if path.exists():
                path.unlink()
        shutil.rmtree(Path("/opt") / name)
    subprocess.run(["systemctl", "daemon-reload"], check=True)
    subprocess.run(["nginx", "-t"], check=True)
    subprocess.run(["systemctl", "reload", "nginx"], check=True)
    (backup / "RESTORE.md").write_text("Restore projects/<name> to /opt/<name> and config/etc files to their original /etc paths. Then run systemctl daemon-reload, nginx -t, systemctl reload nginx and restore service enabled/active state using service-state.json.\n")
    print(json.dumps({"retired": list(PROJECTS), "backup": str(backup), "preserved": "deepseek-harness"}), flush=True)


if __name__ == "__main__":
    main()
