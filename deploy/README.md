# 独立部署与迁移

运行 `bash deploy/deploy.sh`。默认沿用原服务器，可用 `DEPLOY_HOST`、`DEPLOY_PORT` 覆盖。脚本只同步本项目，先构建 Go 服务，再备份、切换与健康检查。

## 从统一网站部署拆分

1. 先部署本项目。新程序位于 `/opt/knowledge-graph/current`，由 `knowledge-graph.service` 管理，仍监听公开只读端口 `8180` 和认证工作区端口 `8181`。
2. 文档和 SQLite 继续使用 `/var/lib/henryweb/documents` 与 `/var/lib/henryweb/content.db`。首次切换从 `/etc/henryweb/knowledge.env`（或原 DevGraph 环境文件）复制已有凭据到 `/etc/knowledge-graph/auth.env`，不重置账号，不重发文档，不复制数据到源码仓库。
3. 旧 `henryweb-api` 写入进程在最终备份前停止，新服务接管同一库，避免两个进程同时写。失败时恢复原程序、数据与 Nginx 配置。
4. API 和知识域名配置迁到 `/etc/nginx/conf.d/zz-knowledge-graph.conf`；从旧的统一配置中只移除这两个域名，保留博客虚拟主机。原 IP 的 `/graph/` 配置不改动，端口和 Cookie 路径兼容。
5. 再独立部署网站项目。网站只更新 Next.js，不再停止内容服务或接触数据库。

域名和证书继续使用现有安装。全新服务器需先按 `nginx.conf` 准备 TLS、设置 `/etc/knowledge-graph/auth.env` 并配置已有所有者的 `DEVGRAPH_USERNAME`、`DEVGRAPH_PASSWORD_HASH`、`DEVGRAPH_PASSWORD_SALT`。账号配置不要提交。

## 日常维护

- 服务状态：`systemctl status knowledge-graph`；日志：`journalctl -u knowledge-graph`。
- 维护配置为 `/etc/knowledge-graph/api.env` 与 `auth.env`，后续部署保留配置。
- Markdown、`document.json` 和 SQLite 一起备份。发布快照、历史与会话都属于持久数据。
- 使用 `contentctl` 之前先停止 `knowledge-graph`。维护结束后重新启动；进程锁阻止第二个写入进程。
- 发布和撤回通常在工作区完成；`contentctl --publish-all` 仅用于明确需要批量公开时，普通部署不调用它。
- 迁移备份位于 `/opt/knowledge-graph/backups/<release>`，目录仅 root 可读，不进入 Git 或同步包。
