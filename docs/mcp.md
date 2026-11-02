# Neurix 知识库接入

知识图谱通过只读 MCP 为 Neurix 提供当前 Markdown 笔记，包括私有草稿、未在图谱展示的文档和关联关系。通过现有 SSH 身份验证访问服务器，进程使用 `henryweb-content` 的文件读取权限；不需要网页密码、会话 Cookie 或新的公开接口。笔记内容会随工具结果提供给当前会话所选模型。

## 安装与连接

```sh
npm run build
bash deploy/deploy-mcp.sh
```

该脚本仅安装独立的 `knowledge-mcp` 可执行文件，以内容账户预检读取权限，校验 SHA-256 后切换 `/opt/knowledge-graph/mcp/current`。不会重启内容服务、打开数据库、迁移文件或改动发布状态。旧可执行文件保留在 `mcp/releases/`，上一版本路径记录在权限为 `0700` 的 `mcp/backups/<release>/previous-release` 中。回滚时将 `mcp/current` 原子切回记录的目录。

将 [`deploy/neurix-knowledge.cordis.yml`](../deploy/neurix-knowledge.cordis.yml) 中的插入项合并到 `~/.dsh/cordis.patch.yml`，可供本机所有 Neurix profile 使用。只希望一个 profile 使用时，合并到 `~/.dsh/profiles/<name>/cordis.patch.yml`。保留文件中原有设置；自定义 `DSH_HOME` 时使用对应目录。连接依赖当前系统用户已有的 SSH 配置和密钥。

重启对应 Neurix profile，或等待其配置热重载，随后新建会话并输入：

> 搜索我的知识库，找出与部署有关的笔记，读取相关段落并附上来源。

应看到 `mcp__knowledge__search_documents`、`mcp__knowledge__read_document` 和 `mcp__knowledge__related_documents`。可用 `pnpm dsh --profile web --dump-config` 检查 Web profile 的配置合成；该命令只检查配置，不验证 SSH 连通性。Desktop profile 由 Electron 应用管理，可在桌面应用的插件设置中查看连接。

## 工具

| 工具 | 用途 |
| --- | --- |
| `search_documents` | 搜索标题、摘要、标签、分类和正文，返回排序后的摘要、匹配片段、UUID 和来源链接；空查询列出最近笔记。 |
| `read_document` | 按 UUID 读取最新 Markdown；默认最多 8000 个 Unicode 字符，可使用 `next_offset` 续读。 |
| `related_documents` | 查询出链、反向链接和未解析引用，忽略代码中的链接；结果支持分页。 |

三个工具均支持 `locale: zh-CN / zh-TW / en`，默认简体中文。不同语言共享原文 UUID；缺少译文时按工作区规则回退，`content_locale` 和 `available_locales` 明确标示实际语言。搜索与关联查询默认返回 10 条，`limit` 最多 30，`offset` 用于翻页。

搜索使用不区分大小写的关键词匹配。空格分隔的词必须全部命中，中文可直接使用短语；目前没有同义词或向量检索。来源链接指向现有的私有文档 JSON 接口，浏览器登录知识工作区后可查看。每次调用都重新读取文件；结果中的 `version` 可用来识别笔记是否更新。

## 本地目录

本地资料需使用本项目的 `<UUID>/<locale>.md` 格式。可直接运行：

```sh
backend/bin/knowledge-mcp --root /absolute/path/to/documents --check
backend/bin/knowledge-mcp --root /absolute/path/to/documents --workspace-url http://127.0.0.1:8181
```

在 Neurix 的 MCP 配置中把 `command` 换为本地可执行文件的绝对路径，`args` 换为上述参数。`--check` 只用于预检，不应出现在正式 MCP 参数中。

MCP 仅通过标准输入输出通信，不监听网络端口。它没有写入、删除、发布或会话管理工具，也不加载 SQLite；内容目录内的符号链接会被拒绝。Markdown 中的文本作为资料处理，不作为操作指令。原有公开 API 的发布、语言与隐私边界保持不变。
