#!/usr/bin/env python3
"""Issue the IP certificate and configure automatic renewal for this server."""

import shutil
import subprocess
from datetime import datetime, timezone
from pathlib import Path


def run(*command):
    return subprocess.run(command, check=True, text=True)


def main():
    certbot = "/opt/devgraph-certbot/bin/certbot"
    run(certbot, "certonly", "--non-interactive", "--agree-tos",
        "--register-unsafely-without-email", "--preferred-profile", "shortlived",
        "--webroot", "--webroot-path", "/var/www/acme",
        "--ip-address", "101.34.211.31", "--cert-name", "devgraph-ip",
        "--deploy-hook", "/usr/sbin/nginx -t && /usr/bin/systemctl reload nginx")
    directory = Path("/etc/letsencrypt/live/devgraph-ip")
    if not all((directory / name).exists() for name in ["fullchain.pem", "privkey.pem"]):
        raise RuntimeError("Certificate files were not created.")

    config = Path("/etc/nginx/conf.d/tactics-harness.conf")
    original = config.read_text()
    backup = Path("/opt/backups/devgraph/tls-setup")
    backup.mkdir(parents=True, exist_ok=True, mode=0o700)
    shutil.copy2(config, backup / ("before-trusted-tls-" + datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ") + ".conf"))
    changed = original.replace("ssl_certificate /etc/nginx/dsh-certs/cert.pem;", "ssl_certificate /etc/letsencrypt/live/devgraph-ip/fullchain.pem;")
    changed = changed.replace("ssl_certificate_key /etc/nginx/dsh-certs/key.pem;", "ssl_certificate_key /etc/letsencrypt/live/devgraph-ip/privkey.pem;")
    if "/etc/letsencrypt/live/devgraph-ip/fullchain.pem" not in changed:
        raise RuntimeError("Expected certificate directive was not found; configuration unchanged.")
    config.write_text(changed)
    try:
        run("nginx", "-t")
        run("systemctl", "reload", "nginx")
    except Exception:
        config.write_text(original)
        run("nginx", "-t")
        run("systemctl", "reload", "nginx")
        raise

    Path("/etc/systemd/system/devgraph-certbot-renew.service").write_text('''[Unit]
Description=Renew DevGraph IP HTTPS certificate
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/opt/devgraph-certbot/bin/certbot renew --quiet --cert-name devgraph-ip
''')
    Path("/etc/systemd/system/devgraph-certbot-renew.timer").write_text('''[Unit]
Description=Check DevGraph certificate renewal every eight hours

[Timer]
OnBootSec=10min
OnUnitActiveSec=8h
RandomizedDelaySec=30min
Persistent=true

[Install]
WantedBy=timers.target
''')
    run("systemctl", "daemon-reload")
    run("systemctl", "enable", "--now", "devgraph-certbot-renew.timer")
    run("openssl", "x509", "-in", str(directory / "fullchain.pem"), "-noout", "-issuer", "-dates", "-ext", "subjectAltName")
    print("Trusted HTTPS and certificate renewal timer configured.", flush=True)


if __name__ == "__main__":
    main()
