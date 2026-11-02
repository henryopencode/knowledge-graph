#!/usr/bin/env bash
set -euo pipefail
source_dir=$(cd "$(dirname "$0")/.." && pwd)
deploy_host=${DEPLOY_HOST:-root@101.34.211.31}
deploy_port=${DEPLOY_PORT:-22}
release_id=$(date -u +%Y%m%dT%H%M%SZ)
remote_source="/opt/knowledge-graph/releases/$release_id"
if ! [[ "$deploy_port" =~ ^[0-9]+$ ]]; then echo 'Invalid DEPLOY_PORT' >&2; exit 1; fi
ssh -p "$deploy_port" "$deploy_host" "install -d -m 0755 '$remote_source'"
rsync -az --exclude-from="$source_dir/deploy/rsync-excludes.txt" -e "ssh -p $deploy_port" "$source_dir/" "$deploy_host:$remote_source/"
ssh -p "$deploy_port" "$deploy_host" "bash '$remote_source/deploy/activate.sh'"
printf '\nDeployed https://knowledge.henryweb.site\n'
