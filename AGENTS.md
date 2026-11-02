# Working rules

1. Avoid unnecessary reading and redundant steps; keep responses concise.
2. Preserve current behavior, compatibility and the existing owner's credentials.
3. Keep interfaces free of redundant slogans; retain operation, error and accessibility text.

# Boundaries

- `web/`: authenticated graph UI, with the migrated working-tree improvements preserved.
- `backend/internal/content`: Markdown validation, atomic writes, revisions, publication snapshots and attachments.
- `backend/internal/auth`: owner credentials and host-only sessions. Never add credentials to tracked files.
- `contracts/openapi.json`: the content API, including the public routes consumed by the independently deployed blog. Preserve field names, IDs, error codes and the three locales `zh-CN`, `zh-TW`, `en`.
- `content/documents/`: explicitly public seed examples only. `vault/`, `var/`, local learning material, credentials, databases, backups and legacy source archives stay ignored.
- Do not depend on the website repository's source or build directories.

# Verification

Run `npm run check`, `go test -tags=nomsgpack ./...` from `backend`, `npm run build`, and `bash -n deploy/*.sh`. Visibility, publication, translation identity and authentication changes require behavioral tests.

# Deployment

`bash deploy/deploy.sh` activates only this project's two-listener content service. Preserve the existing IP `/graph/` route, host-only cookies, credentials and publication state. Keep migration backups root-only on the server. The website is deployed separately after the first split.
