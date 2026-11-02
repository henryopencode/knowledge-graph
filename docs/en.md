# DevGraph: development and content guide

[简体中文](zh-CN.md) · [繁體中文](zh-TW.md) · English

## Architecture

This repository owns the private graph interface in `web/` and the Go/Gin content service. The Next.js blog is an independent project that reads published articles through the public API. The service exposes separate listeners: `127.0.0.1:8180` for public read-only APIs and `127.0.0.1:8181` for the authenticated workspace. The public listener never registers private routes.

`henryweb.site/{locale}` serves the blog, `api.henryweb.site` serves public APIs, and `knowledge.henryweb.site` serves the workspace. Supported locales are `zh-CN`, `zh-TW`, and `en`. Old blog links redirect to Simplified Chinese. The old IP `/graph/` entry remains available.

## Documents

Production documents live outside the repository under `/var/lib/henryweb/documents/<UUID>/<locale>.md`. One stable UUID identifies a logical document across languages. Titles and slugs can change without changing its identity. The repository contains only public seed examples.

```yaml
---
schema_version: 1
id: "67e68cef-91a1-4be3-b07f-fb5c0aa95eed"
locale: en
title: "A programming note"
summary: "A short summary."
slug: a-programming-note
kind: concept
category: notes
tags: [go]
channels:
  blog: false
  graph: true
created: "2026-10-26T00:00:00Z"
updated: "2026-10-26T00:00:00Z"
---
```

The editor creates metadata and IDs. `kind` is one of `project`, `incident`, `decision`, or `concept`. Category IDs and API enums are stable across locales. `notes` and `site` have localized display labels. Custom category labels are shown as authored.

New documents are private and included in the graph by default. The blog and graph flags are independent. `<UUID>/document.json` stores the authoritative shared blog policy (`schema_version: 1`, `blog: true/false`). Markdown `channels.blog` remains a compatibility field; it cannot override the shared policy. Include this JSON file when backing up or moving a document. Documents excluded from the graph remain in the file tree and document list. Import creates private drafts and never publishes incoming Markdown automatically.

Link documents with `[[doc:UUID|label]]`. Legacy `[[title]]` links still work when the title uniquely identifies a document. Source code and fenced examples never create graph edges. A translated document remains one graph node; when a translation is missing, the workspace labels the available source language.

## Publishing and translation

Saving changes only the working Markdown version. Enable blog publication, save, preview the current revision, then publish the whole document. Publishing, withdrawal, and disabling blog visibility apply to all three interface languages together. A single transaction saves every existing language variant and its attachments. The public blog reads a reviewed snapshot from SQLite. Editing a published document does not modify that snapshot. Withdrawal removes its public version immediately. Disabling blog visibility also prevents public reads. Search, feeds, the sitemap, translations, and attachments all use the same public boundary and `no-store` responses.

Public responses expose explicit fields, not raw frontmatter or private graph metadata. Private link targets and aliases are replaced by a neutral label. Private local paths are not exposed. Referenced `assets/` attachments are copied into the publication snapshot; subsequent edits remain private until republishing. HTML is sanitized. Whole-document visibility is the supported boundary; put confidential additions in a separate private document.

Each language has its own Markdown title, summary, body, slug, and editing revision. Publication status is shared across the document. The publishing area has no language selector; choose a language inside the editor to author translations. A missing published translation displays the published original with a source-language notice, never a working draft. `locale` is the requested interface language, `content_locale` identifies the actual content, and `available_locales` lists real published translations. Every `translations` entry is a usable route, including fallback routes. `source_locale` and `source_revision` associate a translation with its source. Source edits mark that translation as potentially outdated. No automated replacement of private or authored content occurs. The language selector changes the interface, not the stored content.

## API and localization

The machine-readable contract is [`contracts/openapi.json`](../contracts/openapi.json). Generate TypeScript types with `npm run api:generate`. Public APIs use `/api/v1/public/posts`, `/api/v1/public/posts/{slug}`, `/api/v1/site`, and `/api/v1/projects`.

Preview, publish, and withdraw accept `version: Document.publication_version`, an aggregate token covering all language drafts and shared visibility. Their optional `locale` selects preview/response content only. An edit to any language invalidates an older preview token. Editing uses the per-Markdown `version`; a visibility change can additionally send `publication_version` to guard against concurrent edits.

Private operations are hosted only on the workspace domain: `/api/v1/private/documents`, document read/update/delete, and `/{id}/preview`, `/{id}/publish`, `/{id}/withdraw`. Mutations require a host-only HttpOnly session cookie, an allowed Origin, and `X-DevGraph-Request: 1`. Legacy graph routes remain as adapters to the same store.

Language precedence is an explicit `locale` query parameter, language preference cookie, then `Accept-Language`. Responses include `Content-Language`; caches vary by language and origin. Error codes and field names never change by language. Modern errors contain `error.code`, `error.message`, and `request_id`. Legacy graph errors retain the older shape. Backend messages are in `backend/internal/i18n/messages`; blog messages are in `src/messages`. Catalog and placeholder consistency is checked automatically.

## Local development and verification

Use Go 1.25+ and Node.js 20.9+. There are no frontend build dependencies. Run `go run -tags=nomsgpack ./cmd/server` from `backend`; static assets default to `../web`.

Configure `DEVGRAPH_USERNAME`, `DEVGRAPH_PASSWORD_HASH` and `DEVGRAPH_PASSWORD_SALT` first. Credentials use PBKDF2-HMAC-SHA256, 260000 iterations, a 16-byte hex salt and a 32-byte hex digest. Production retains the existing owner account; use `COOKIE_SECURE=false` only for local HTTP. Never commit credentials.

Local storage defaults to `backend/var/documents` and `backend/var/content.db`. Override `CONTENT_ROOT`, `CONTENT_DB` and `KNOWLEDGE_ASSETS` as needed. The public listener remains on port 8180 and the authenticated workspace on 8181.

```sh
npm run check
npm run test:backend
npm run build
python3 -m unittest discover -s deploy/tests -v
bash -n deploy/*.sh
```

## Deployment, backup and rollback

See the [independent deployment and migration guide](../deploy/README.md). Deploy the knowledge project before the separate website for the initial split. Content and credentials retain their existing paths and formats; runtime data never belongs in either source repository.

For offline maintenance, stop `knowledge-graph.service`, run `backend/bin/contentctl`, then restart the service. Back up Markdown, document.json and SQLite together, keeping backups root-only. Ordinary deployment does not publish all documents.
