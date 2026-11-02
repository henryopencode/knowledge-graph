# DevGraph：开发与内容管理

简体中文 · [繁體中文](zh-TW.md) · [English](en.md)

## 架构

本仓库包含 `web/` 私有图谱界面和 Go/Gin 内容服务；Next.js 公开博客在独立的网站项目中，通过公开 API 读取已发布文章。内容服务有两个独立监听入口：`127.0.0.1:8180` 提供公开只读接口，`127.0.0.1:8181` 提供登录后的工作区。公开入口不注册私有路由。

工作区是需要原有账号密码登录的纯展示界面：左侧按已有标签和分类组织主题，中间阅读正文，右侧查看本页目录、出链和反向链接。主题目录只影响浏览，不改写文档元数据。搜索和筛选使用笔记列表；关联图谱默认显示当前文档的一跳邻居，较大范围先显示 30 篇，可搜索缩小范围或切换全局。`#doc:<UUID>` 可保存当前阅读位置，切换界面语言继续使用同一文档 ID。页面不提供新建、编辑、导入、备份导出、删除、发布或撤回操作；缺失关联只显示状态，不触发创建。内容维护使用现有受认证内容 API 或离线维护流程，公开博客继续读取原有发布快照。

博客地址为 `henryweb.site/{locale}`，公开接口为 `api.henryweb.site`，工作区为 `knowledge.henryweb.site`。支持 `zh-CN`、`zh-TW`、`en`。原博客地址跳转到简体中文版；原 IP 的 `/graph/` 入口继续可用。

## 文档规范

生产文档保存在仓库外的 `/var/lib/henryweb/documents/<UUID>/<locale>.md`。同一篇文档的不同语言共用固定 UUID，改标题或文章地址不会改变文档身份。仓库内仅保存公开示例。

```yaml
---
schema_version: 1
id: "67e68cef-91a1-4be3-b07f-fb5c0aa95eed"
locale: zh-CN
title: "一篇编程笔记"
summary: "一句话简介。"
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

内容服务创建 ID 和元数据。`kind` 使用 `project`、`incident`、`decision`、`concept`。分类标识和接口枚举跨语言保持稳定；`notes`、`site` 提供三语显示名称，自定义分类按作者填写的内容显示。

新文档默认私有并进入图谱。博客与图谱开关独立。`<UUID>/document.json` 保存整篇文档共享的博客开关（`schema_version: 1`、`blog: true/false`）。Markdown 的 `channels.blog` 作为兼容字段，不能覆盖共享配置。备份、搬移文档时须一并保留该 JSON 文件。关闭图谱展示后，文档仍能在主题目录和文档列表中阅读。通过 API 导入始终产生私有草稿，不会自动发布传入的 Markdown。

使用 `[[doc:UUID|显示名称]]` 关联文档。旧的 `[[标题]]` 链接在标题唯一时继续兼容。代码块与行内代码不会生成关联。不同语言在图谱中仍是一个节点，缺少译文时会标明当前显示的原文语言。

## 发布与翻译

保存只更新工作版本。开启博客展示并保存后，先预览当前版本，再发布整篇文档。发布、撤回和关闭博客展示对三种界面语言同时生效；一次事务保存全部已有语言及其附件快照。博客读取 SQLite 中的公开快照。继续编辑已发布文档不会改变公开版本；撤回后公开文章立即停止提供。关闭博客展示也会阻止公开读取。搜索、RSS、站点地图、译文和附件使用相同的公开边界，响应禁止缓存。

公开接口使用明确的数据字段，不返回原始 Frontmatter 和私有图谱元数据。指向私有文档的目标与别名会替换为中性文本；私有路径不会暴露。正文引用的 `assets/` 附件在发布时生成快照，后续修改保持私有，重新发布后才更新。HTML 会经过安全过滤。权限按整篇文档划分，私密补充应单独放在关联的私有文档里。

每种语言独立维护标题、摘要、正文、地址和编辑版本，发布状态由整篇文档共享。发布操作无需选择语言；译文通过内容 API 按语言维护。缺少已发布译文时显示已发布的原文并标明原文语言，不读取草稿。接口的 `locale` 表示请求的界面语言，`content_locale` 表示实际正文语言，`available_locales` 列出真实已发布译文，`translations` 为三种语言提供可访问的地址（包含原文回退）。`source_locale`、`source_revision` 记录译文对应的原文。原文更新后，系统提示译文可能过期，不自动覆盖内容。界面语言切换不会改写文档正文。

## 接口与多语言

接口协议见 [`contracts/openapi.json`](../contracts/openapi.json)，通过 `npm run api:generate` 生成前端 TypeScript 类型。公开接口包括 `/api/v1/public/posts`、`/api/v1/public/posts/{slug}`、`/api/v1/site`、`/api/v1/projects`。

预览、发布和撤回提交 `version: Document.publication_version`，该聚合版本覆盖所有语言草稿及共享开关；可选的 `locale` 只决定预览和响应内容。任一语言改动都会使旧预览版本失效。编辑仍使用单份 Markdown 的 `version`，修改共享开关时可额外提交 `publication_version` 防止并发覆盖。

私有接口仅存在于知识图谱域名：`/api/v1/private/documents`、单文档读写删除以及 `/{id}/preview`、`/{id}/publish`、`/{id}/withdraw`。写操作需要主机独立的 HttpOnly 会话、允许的 Origin 和 `X-DevGraph-Request: 1`。旧图谱接口通过适配层访问同一存储。

语言优先级依次为显式 `locale` 参数、语言偏好 Cookie、`Accept-Language`。响应包含 `Content-Language`，缓存按语言和来源区分。错误码和字段名称保持稳定；新版错误包含 `error.code`、`error.message`、`request_id`，旧图谱接口保留原有错误结构。后端文案在 `backend/internal/i18n/messages`，博客文案在 `src/messages`，检查脚本验证三语键名和占位符一致。

## 本地开发与检查

需要 Go 1.25+ 和 Node.js 20.9+。项目没有前端构建依赖。工作区静态资源在 `web/`，服务从 `backend` 目录运行时默认读取 `../web`。

在 `backend` 运行 `go run -tags=nomsgpack ./cmd/server` 前，配置 `DEVGRAPH_USERNAME`、`DEVGRAPH_PASSWORD_HASH`、`DEVGRAPH_PASSWORD_SALT`。密码使用 PBKDF2-HMAC-SHA256，260000 次迭代，16 字节十六进制盐和 32 字节十六进制摘要。生产环境沿用原账号；只有本地 HTTP 使用 `COOKIE_SECURE=false`。凭据不得提交。

本地默认存储为 `backend/var/documents` 与 `backend/var/content.db`。可用 `CONTENT_ROOT`、`CONTENT_DB` 和 `KNOWLEDGE_ASSETS` 明确配置；工作区默认端口 8181，公开只读端口 8180。

```sh
npm run check
npm run test:backend
npm run build
python3 -m unittest discover -s deploy/tests -v
bash -n deploy/*.sh
```

## 部署、备份与回滚

参见 [独立部署与迁移](../deploy/README.md)。先部署知识项目，再部署独立博客。内容与凭据保持现有路径及格式；不需要把数据库或私有 Markdown 搬入任何仓库。

维护使用 `backend/bin/contentctl`，运行前先停止 `knowledge-graph.service`，维护后重新启动。Markdown、document.json 与 SQLite 需要一起备份，备份保持 root-only。普通部署不执行批量公开。
