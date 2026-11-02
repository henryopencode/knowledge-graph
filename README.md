# DevGraph · 知识图谱

独立的 Markdown 知识库：Go/Gin 内容服务、SQLite 发布快照和需要密码登录的笔记阅读界面。支持简体中文、繁體中文、English。

- [知识工作区](https://knowledge.henryweb.site)
- [公开内容 API](https://api.henryweb.site)
- [开发与数据规范](docs/zh-CN.md)
- [部署与拆分迁移](deploy/README.md)
- [API 协议](contracts/openapi.json)
- [Neurix 知识库接入](docs/mcp.md)

本项目拥有文档、鉴权、图谱、附件和发布流程。博客是独立应用，只通过公开 API 读取已发布快照，不读取私有目录。

## 本地运行

需要 Go 1.25+；Node.js 20.9+ 用于界面脚本与语言检查，无前端构建依赖。

在 `backend` 目录运行 `go run -tags=nomsgpack ./cmd/server`。运行前配置 `DEVGRAPH_USERNAME`、`DEVGRAPH_PASSWORD_HASH`、`DEVGRAPH_PASSWORD_SALT`；本地 HTTP 使用 `COOKIE_SECURE=false`。工作区为 `http://127.0.0.1:8181`，公开只读 API 为 `http://127.0.0.1:8180`。账号、内容目录和数据库配置见开发指南。

```sh
npm run check
npm run test:backend
npm run build
```

## 从统一项目拆出

当前 Go 服务和 `web/` 来自网站项目的最新工作区，保留分类筛选等尚未提交的界面改动。原 Python 版本源码保存在被忽略的 `.legacy-python-20261102/`；原 `vault/`、`.local/` 等本地资料保留在原位并排除出 Git。它们不会自动导入当前生产库。

生产文档、数据库、已发布快照和会话继续使用现有路径及格式，首次部署沿用原账号。不要将私有内容或凭据复制到仓库；提交前检查忽略规则。
