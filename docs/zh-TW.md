# DevGraph：開發與內容管理

[简体中文](zh-CN.md) · 繁體中文 · [English](en.md)

## 架構

本倉庫包含 `web/` 私有圖譜介面和 Go/Gin 內容服務；Next.js 公開部落格是獨立專案，透過公開 API 讀取已發佈文章。內容服務有兩個獨立監聽入口：`127.0.0.1:8180` 提供公開只讀接口，`127.0.0.1:8181` 提供登入後的工作區。公開入口不註冊私有路由。

博客地址為 `henryweb.site/{locale}`，公開接口為 `api.henryweb.site`，工作區為 `knowledge.henryweb.site`。支持 `zh-CN`、`zh-TW`、`en`。原博客地址跳轉到簡體中文版；原 IP 的 `/graph/` 入口繼續可用。

## 文檔規範

生產文檔儲存在倉庫外的 `/var/lib/henryweb/documents/<UUID>/<locale>.md`。同一篇文檔的不同語言共用固定 UUID，改標題或文章地址不會改變文檔身份。倉庫內僅儲存公開示例。

```yaml
---
schema_version: 1
id: "67e68cef-91a1-4be3-b07f-fb5c0aa95eed"
locale: zh-CN
title: "一篇編程筆記"
summary: "一句話簡介。"
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

編輯器自動創建 ID 和元數據。`kind` 使用 `project`、`incident`、`decision`、`concept`。分類標識和接口枚舉跨語言保持穩定；`notes`、`site` 提供三語顯示名稱，自定義分類按作者填寫的內容顯示。

新文檔默認私有並進入圖譜。博客與圖譜開關獨立。`<UUID>/document.json` 儲存整篇文檔共用的博客開關（`schema_version: 1`、`blog: true/false`）。Markdown 的 `channels.blog` 作為相容欄位，不能覆蓋共用設定。備份、搬移文檔時須一併保留該 JSON 檔案。關閉圖譜展示後，文檔仍能在文件樹和文檔列表中編輯。匯入始終產生私有草稿，不會自動發佈傳入的 Markdown。

使用 `[[doc:UUID|顯示名稱]]` 關聯文檔。舊的 `[[標題]]` 鏈接在標題唯一時繼續兼容。代碼塊與行內代碼不會生成關聯。不同語言在圖譜中仍是一個節點，缺少譯文時會標明當前顯示的原文語言。

## 發佈與翻譯

儲存只更新工作版本。開啟博客展示並儲存後，先預覽當前版本，再發佈整篇文檔。發佈、撤回和關閉博客展示對三種介面語言同時生效；一次交易儲存全部已有語言及其附件快照。博客讀取 SQLite 中的公開快照。繼續編輯已發佈文檔不會改變公開版本；撤回後公開文章立即停止提供。關閉博客展示也會阻止公開讀取。搜索、RSS、站點地圖、譯文和附件使用相同的公開邊界，響應禁止緩存。

公開接口使用明確的數據字段，不返回原始 Frontmatter 和私有圖譜元數據。指向私有文檔的目標與別名會替換為中性文本；私有路徑不會暴露。正文引用的 `assets/` 附件在發佈時生成快照，後續修改保持私有，重新發布後才更新。HTML 會經過安全過濾。權限按整篇文檔劃分，私密補充應單獨放在關聯的私有文檔裡。

每種語言獨立維護標題、摘要、正文、地址和編輯版本，發佈狀態由整篇文檔共用。發佈區域無須選擇語言；譯文在編輯器中切換語言後撰寫。缺少已發佈譯文時顯示已發佈的原文並標明原文語言，不讀取草稿。介面的 `locale` 表示請求的介面語言，`content_locale` 表示實際正文語言，`available_locales` 列出真實已發佈譯文，`translations` 為三種語言提供可存取的地址（包含原文回退）。`source_locale`、`source_revision` 記錄譯文對應的原文。原文更新後，系統提示譯文可能過期，不自動覆蓋內容。界面語言切換不會改寫文檔正文。

## 接口與多語言

接口協議見 [`contracts/openapi.json`](../contracts/openapi.json)，通過 `npm run api:generate` 生成前端 TypeScript 類型。公開接口包括 `/api/v1/public/posts`、`/api/v1/public/posts/{slug}`、`/api/v1/site`、`/api/v1/projects`。

預覽、發佈和撤回提交 `version: Document.publication_version`，該聚合版本涵蓋所有語言草稿及共用開關；可選的 `locale` 只決定預覽和回應內容。任一語言改動都會使舊預覽版本失效。編輯仍使用單份 Markdown 的 `version`，修改共用開關時可額外提交 `publication_version` 防止並行覆蓋。

私有接口僅存在於知識圖譜域名：`/api/v1/private/documents`、單文檔讀寫刪除以及 `/{id}/preview`、`/{id}/publish`、`/{id}/withdraw`。寫操作需要主機獨立的 HttpOnly 會話、允許的 Origin 和 `X-DevGraph-Request: 1`。舊圖譜接口通過適配層訪問同一存儲。

語言優先級依次為顯式 `locale` 參數、語言偏好 Cookie、`Accept-Language`。響應包含 `Content-Language`，緩存按語言和來源區分。錯誤碼和字段名稱保持穩定；新版錯誤包含 `error.code`、`error.message`、`request_id`，舊圖譜接口保留原有錯誤結構。後端文案在 `backend/internal/i18n/messages`，博客文案在 `src/messages`，檢查腳本驗證三語鍵名和佔位符一致。

## 本機開發與檢查

需要 Go 1.25+ 和 Node.js 20.9+，沒有前端建置相依套件。從 `backend` 執行 `go run -tags=nomsgpack ./cmd/server`，預設靜態資源為 `../web`。

啟動前設定 `DEVGRAPH_USERNAME`、`DEVGRAPH_PASSWORD_HASH`、`DEVGRAPH_PASSWORD_SALT`。使用 PBKDF2-HMAC-SHA256，260000 次迭代、16 位元組十六進位鹽與 32 位元組摘要。正式環境沿用原帳號；僅本機 HTTP 設定 `COOKIE_SECURE=false`。憑證不得提交。

本機資料預設位於 `backend/var/documents`、`backend/var/content.db`，可透過 `CONTENT_ROOT`、`CONTENT_DB` 與 `KNOWLEDGE_ASSETS` 覆寫。工作區端口為 8181，公開唯讀端口為 8180。

```sh
npm run check
npm run test:backend
npm run build
python3 -m unittest discover -s deploy/tests -v
bash -n deploy/*.sh
```

## 部署、備份與回復

參見 [獨立部署與遷移](../deploy/README.md)。先部署知識專案，再部署獨立部落格。資料與帳號保留既有路徑及格式，不進入程式碼倉庫。

離線維護使用 `backend/bin/contentctl`，執行前停止 `knowledge-graph.service`，維護後重新啟動。Markdown、document.json 和 SQLite 一起備份，備份僅 root 可讀。一般部署不會批次公開文件。
