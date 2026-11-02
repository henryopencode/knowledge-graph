#!/usr/bin/env bash
# Install only the SSH/stdin knowledge reader; the content service stays running.
set -euo pipefail
source_dir=$(cd "$(dirname "$0")/.." && pwd)
deploy_host=${DEPLOY_HOST:-root@101.34.211.31}
deploy_port=${DEPLOY_PORT:-22}
mcp_arch=${MCP_GOARCH:-amd64}
if ! [[ "$deploy_port" =~ ^[0-9]+$ ]]; then echo 'Invalid DEPLOY_PORT' >&2; exit 1; fi
case "$mcp_arch" in amd64|arm64) ;; *) echo 'MCP_GOARCH must be amd64 or arm64' >&2; exit 1 ;; esac
build_dir=$(mktemp -d)
trap 'rm -rf "$build_dir"' EXIT
(
    cd "$source_dir/backend"
    GOOS=linux GOARCH="$mcp_arch" CGO_ENABLED=0 go build -trimpath -tags=nomsgpack -ldflags='-s -w' -o "$build_dir/knowledge-mcp" ./cmd/knowledge-mcp
)
checksum=$(shasum -a 256 "$build_dir/knowledge-mcp" | cut -d ' ' -f 1)
release_id="$(date -u +%Y%m%dT%H%M%SZ)-${checksum:0:12}"
remote_base=/opt/knowledge-graph/mcp
remote_release="$remote_base/releases/$release_id"
ssh -p "$deploy_port" "$deploy_host" "install -d -m 0755 '$remote_release'"
scp -P "$deploy_port" "$build_dir/knowledge-mcp" "$deploy_host:$remote_release/knowledge-mcp.upload"
ssh -p "$deploy_port" "$deploy_host" "bash -s -- '$remote_base' '$remote_release' '$release_id' '$checksum'" <<'REMOTE'
set -euo pipefail
mcp_base=$1
mcp_release=$2
mcp_release_id=$3
expected_checksum=$4
test "$(id -u)" = 0
actual_checksum=$(sha256sum "$mcp_release/knowledge-mcp.upload" | cut -d ' ' -f 1)
test "$actual_checksum" = "$expected_checksum"
test ! -e "$mcp_release/knowledge-mcp"
mv "$mcp_release/knowledge-mcp.upload" "$mcp_release/knowledge-mcp"
chmod 0755 "$mcp_release/knowledge-mcp"
# Preflight with the content owner's filesystem permissions, without SQLite.
runuser -u henryweb-content -- "$mcp_release/knowledge-mcp" --root /var/lib/henryweb/documents --check
install -d -m 0700 "$mcp_base/backups/$mcp_release_id"
readlink -f "$mcp_base/current" > "$mcp_base/backups/$mcp_release_id/previous-release" || true
ln -s "$mcp_release" "$mcp_base/current.$mcp_release_id"
mv -Tf "$mcp_base/current.$mcp_release_id" "$mcp_base/current"
printf 'Activated read-only knowledge MCP: %s\n' "$mcp_release_id"
REMOTE
