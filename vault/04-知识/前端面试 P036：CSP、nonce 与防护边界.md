---
title: "前端面试 P036：CSP、nonce 与防护边界"
kind: "concept"
tags: ["前端面试", "浏览器", "CSP", "XSS", "安全"]
created: "2026-10-26T09:13:34Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "6feebebe-88d9-5cfb-a5b3-c88d4f471647"
---

来源：[方格前端讲面试 · 36.CSP（内容安全策略）原理介绍](https://www.bilibili.com/video/BV14Ttw6WEkt?p=36)，10 分 12 秒。

回看定位：[01:08 策略机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=36&t=68) · [03:03 default-src](https://www.bilibili.com/video/BV14Ttw6WEkt?p=36&t=183) · [07:18 nonce](https://www.bilibili.com/video/BV14Ttw6WEkt?p=36&t=438) · [08:30 报告](https://www.bilibili.com/video/BV14Ttw6WEkt?p=36&t=510)

## CSP 控制什么

Content Security Policy 是浏览器依据页面响应策略，限制脚本、样式、图片、连接等能力的机制。它能降低部分注入攻击风险，但不能替代输出编码、可靠的 HTML 清理、权限校验和安全 DOM 操作。

课程演示恶意脚本在有、无 CSP 的页面中产生不同结果；应理解为该演示策略阻止了那次执行，不能推断任何配置都能防住所有 XSS。

## 指令、来源与回退

```http
Content-Security-Policy: default-src 'self'; script-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'
```

这是说明结构的示例，真实项目需要依据资源来源和嵌入需求配置。self 指同源，比较协议、主机、有效端口，不只是域名字符串相同。

视频称 default-src 管理“所有资源”，需要补充：它给一组 fetch 指令提供回退，并不替代 base-uri、frame-ancestors、form-action 等所有指令。script-src 等显式规则也有自己的作用范围。

## nonce 不是固定密码

服务器为每个响应生成足够不可预测的 nonce，同时写入策略中的 nonce-source 和可信 script 标签的 nonce 属性。允许的脚本要匹配这份策略；不能把示例字符串长期硬编码复用。

hash 也可授权特定内联内容，但内容变化需要同步更新。unsafe-inline、unsafe-eval 等宽松选项会改变防护强度，不能仅因控制台错误消失就宣称配置完成。

## 检查与渐进部署

在 Network 检查主文档响应头，结合控制台与违规报告确认具体策略效果。Content-Security-Policy-Report-Only 主要报告，不代表已经阻止执行；多个策略通常同时约束，不能假定后一个放宽前一个。

report-to 使用报告组，需相应 Reporting-Endpoints 等配置；旧 report-uri 存在兼容与弃用问题，部署前核对目标浏览器。meta 策略也不能覆盖响应头支持的全部能力。

## 纠错与证据范围

视频的 nonce、脚本来源限制思路有用。但“超过 90% 网站裸奔”“能阻止绝大多数注入”没有在课程提供可核查统计依据，这里不把这些数字与绝对防护说法当作事实。

依据：[MDN：CSP](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CSP)。关联：[[前端面试 P039：同源策略、CORS 与请求是否发出]]。
