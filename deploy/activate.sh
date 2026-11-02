#!/usr/bin/env bash
set -euo pipefail
source_dir=$(cd "$(dirname "$0")/.." && pwd)
release_id=$(basename "$source_dir")
app_root=/opt/knowledge-graph
runtime="$app_root/runtime-releases/$release_id"
backup="$app_root/backups/$release_id"
export PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
test "$(id -u)" = 0

# Build before stopping the current writer. No website checkout is needed.
cd "$source_dir/backend"
install -d "$runtime/backend" "$runtime/web" "$backup"
chmod 0700 "$backup"
go build -trimpath -tags=nomsgpack -ldflags='-s -w' -o "$runtime/backend/knowledge-graph" ./cmd/server
go build -trimpath -tags=nomsgpack -ldflags='-s -w' -o "$runtime/backend/contentctl" ./cmd/contentctl
cp -a "$source_dir/web/." "$runtime/web/"

id henryweb-content >/dev/null 2>&1 || useradd --system --home-dir /var/lib/henryweb --shell /sbin/nologin henryweb-content
install -d -m 0700 -o henryweb-content -g henryweb-content /var/lib/henryweb
install -d -m 0755 /etc/knowledge-graph
if [ ! -f /etc/knowledge-graph/auth.env ]; then
    if [ -f /etc/henryweb/knowledge.env ]; then
        install -m 0640 -o root -g henryweb-content /etc/henryweb/knowledge.env /etc/knowledge-graph/auth.env
    elif [ -f /etc/devgraph/runtime.env ]; then
        install -m 0640 -o root -g henryweb-content /etc/devgraph/runtime.env /etc/knowledge-graph/auth.env
    else
        echo 'Configure the existing owner credentials in /etc/knowledge-graph/auth.env first.' >&2
        exit 1
    fi
fi
if [ ! -f /etc/knowledge-graph/api.env ]; then
    install -m 0640 -o root -g henryweb-content "$source_dir/deploy/api.env.example" /etc/knowledge-graph/api.env
fi
# Retain the store and account paths from the unified installation.
if ! grep -qx 'CONTENT_ROOT=/var/lib/henryweb/documents' /etc/knowledge-graph/api.env ||
   ! grep -qx 'CONTENT_DB=/var/lib/henryweb/content.db' /etc/knowledge-graph/api.env; then
    echo 'This activation script expects the existing /var/lib/henryweb content paths.' >&2
    exit 1
fi
previous=$(readlink -f "$app_root/current" || true)
for service in knowledge-graph henryweb-api devgraph; do
    systemctl is-active "$service" > "$backup/$service-state.txt" || true
done
for path in /etc/systemd/system/knowledge-graph.service /etc/nginx/conf.d/zz-knowledge-graph.conf /etc/nginx/conf.d/zz-henryweb.conf; do
    if [ -f "$path" ]; then cp -a "$path" "$backup/$(basename "$path")"; fi
done
cutover=false
store_saved=false
rollback() {
    if [ "$cutover" != true ]; then return; fi
    systemctl stop knowledge-graph || true
    if [ "$store_saved" = true ]; then
        mv /var/lib/henryweb "$backup/failed-store"
        tar -C /var/lib -xzf "$backup/content-store.tar.gz"
    fi
    if [ -n "$previous" ]; then
        ln -s "$previous" "$app_root/current.rollback"
        mv -Tf "$app_root/current.rollback" "$app_root/current"
    else
        rm -f "$app_root/current"
    fi
    for path in /etc/systemd/system/knowledge-graph.service /etc/nginx/conf.d/zz-knowledge-graph.conf; do
        if [ -f "$backup/$(basename "$path")" ]; then
            cp -a "$backup/$(basename "$path")" "$path"
        else
            rm -f "$path"
        fi
    done
    if [ -f "$backup/zz-henryweb.conf" ]; then cp -a "$backup/zz-henryweb.conf" /etc/nginx/conf.d/zz-henryweb.conf; fi
    systemctl daemon-reload
    for service in knowledge-graph henryweb-api devgraph; do
        if grep -qx active "$backup/$service-state.txt"; then systemctl start "$service"; fi
    done
    nginx -t && systemctl reload nginx
    echo 'Restored previous content service and publication state. Backups remain root-only.' >&2
}
trap rollback ERR
cutover=true
for service in knowledge-graph henryweb-api devgraph; do
    if grep -qx active "$backup/$service-state.txt"; then systemctl stop "$service"; fi
done
tar -C /var/lib -czf "$backup/content-store.tar.gz" henryweb
store_saved=true
runuser -u henryweb-content -- "$runtime/backend/contentctl" \
    --root /var/lib/henryweb/documents --db /var/lib/henryweb/content.db
install -m 0644 "$source_dir/deploy/knowledge-graph.service" /etc/systemd/system/knowledge-graph.service
ln -s "$runtime" "$app_root/current.new"
mv -Tf "$app_root/current.new" "$app_root/current"
systemctl daemon-reload
systemctl start knowledge-graph
healthy=false
for attempt in $(seq 1 30); do
    if curl -fsS http://127.0.0.1:8180/healthz >/dev/null 2>&1 &&
       curl -fsS http://127.0.0.1:8181/healthz >/dev/null 2>&1; then healthy=true; break; fi
    sleep 1
done
test "$healthy" = true
# The IP virtual host continues to proxy /graph/ to port 8181 unchanged.
if [ -f /etc/nginx/conf.d/zz-henryweb.conf ]; then
    python3 "$source_dir/deploy/split-nginx.py" /etc/nginx/conf.d/zz-henryweb.conf
fi
install -m 0644 "$source_dir/deploy/nginx.conf" /etc/nginx/conf.d/zz-knowledge-graph.conf
nginx -t
systemctl reload nginx
curl --noproxy '*' -fsS --resolve knowledge.henryweb.site:443:127.0.0.1 'https://knowledge.henryweb.site/?locale=zh-TW' >/dev/null
curl --noproxy '*' -fsS --resolve api.henryweb.site:443:127.0.0.1 https://api.henryweb.site/healthz >/dev/null
systemctl enable knowledge-graph
systemctl disable henryweb-api devgraph >/dev/null 2>&1 || true
cutover=false
trap - ERR
printf 'Activated independent knowledge release: %s\n' "$release_id"
