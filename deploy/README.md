# Server deployment

Runtime requirements: Linux, Python 3.10+, Nginx, systemd. No third-party Python dependencies are required by DevGraph.

Keep application releases separate from persistent data:

- `/opt/devgraph/releases/<release>/`: `server.py`, `web/`, `templates/`.
- `/opt/devgraph/current`: symlink to the active release.
- `/var/lib/devgraph/vault/`: persistent Markdown notes, owned by the `devgraph` service account.
- `/etc/devgraph/runtime.env`: the explicit public origin, login username, PBKDF2 password hash and salt, and the session cookie path. Keep it readable only by root.
- `/etc/devgraph/access.json`: initial login information, readable only by root; keep it out of the knowledge vault and source packages.

The application still listens only on `127.0.0.1`; Nginx is the public entry point. `--public-origin` permits that exact Host and browser Origin; it does not provide authentication by itself. The installation script enables the built-in login using `DEVGRAPH_AUTH_ENABLED=1`, `DEVGRAPH_USERNAME`, `DEVGRAPH_PASSWORD_HASH`, `DEVGRAPH_PASSWORD_SALT`, and `DEVGRAPH_COOKIE_PATH`. Sessions are random, revocable, expire after 12 hours, and are invalidated on service restart. Cookies require HTTPS and use HttpOnly, Secure, and SameSite=Strict. Failed logins are rate limited.

For this server, `install_server.py` mounts the workspace at `https://101.34.211.31/graph/`, with Nginx authentication disabled only in that location because the application enforces its own login. Other routes retain their existing access controls. The standalone Nginx example alternatively shows a Basic Authentication gate; use one of these access-control methods before exposing the API.

Before replacing an existing application, back up its data and deployment configuration, record its service names and enabled state, and preserve a documented restore path. Do not remove unrelated projects or shared databases.

For updates, copy only code into a new release, verify it, change the `current` symlink, and restart `devgraph`. Do not copy the starter `vault/` over the existing server vault. Roll back by restoring the previous release, its saved Nginx configuration and runtime environment together, then restart the service and validate Nginx before reloading it.

Back up `/var/lib/devgraph/vault/` independently from releases. The UI export also includes recoverable notes from `.trash/`.
