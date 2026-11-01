---
title: "前端面试 P039：同源策略、CORS 与请求是否发出"
kind: "concept"
tags: ["前端面试", "浏览器", "同源策略", "CORS", "跨域"]
created: "2026-10-26T09:13:34Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "269c94c9-5bc8-5fed-825f-4a6762c0a7e6"
---

来源：[方格前端讲面试 · 39.浏览器的同源策略](https://www.bilibili.com/video/BV14Ttw6WEkt?p=39)，10 分 26 秒。

回看定位：[01:32 同源三要素](https://www.bilibili.com/video/BV14Ttw6WEkt?p=39&t=92) · [03:20 fetch 示例](https://www.bilibili.com/video/BV14Ttw6WEkt?p=39&t=200) · [05:23 CORS](https://www.bilibili.com/video/BV14Ttw6WEkt?p=39&t=323) · [08:24 代理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=39&t=504)

## 同源判断

通常比较协议、主机、有效端口。路径不同不影响同源，子域名不同通常不同源，https 默认端口与显式 :443 等价。

```js
const a = new URL('https://example.com/page1');
const b = new URL('https://example.com:443/page2');
const c = new URL('https://api.example.com/page1');
console.log(a.origin === b.origin); // true
console.log(a.origin === c.origin); // false
```

同源策略限制跨源读取、DOM 访问和相应存储隔离，不是简单禁止任何跨源通信。图片、脚本、表单、导航等有各自的加载或写入规则。

## 纠正：跨域报错不等于请求没发送

简单 CORS 请求通常可以先发出，响应是否交给脚本读取还要看服务器许可。需要预检的请求会先发送 OPTIONS，预检未获允许时通常不继续发送实际请求。

因此 fetch 报 CORS 错误，不能单独证明服务端没收到请求或没执行操作。CORS 也不是 CSRF 防护、身份认证或授权机制。

## CORS 谁配置什么

服务器在响应头发送 Access-Control-Allow-Origin 等许可信息，不是往请求头里加这份许可。凭据场景通常要求客户端 credentials 配置、服务端允许凭据并明确来源，不能用通配符任意放行。实际 Cookie 还受 SameSite 与浏览器第三方策略影响。

动态按来源返回时应正确处理 Vary: Origin、允许的请求头和方法等条件，不要无校验反射任何 Origin。这里只记录机制，不替具体项目直接改策略。

## JSONP、代理与跨窗通信

JSONP 利用经典脚本加载并执行回调，响应是可执行脚本，存在信任与方法限制，不是通用 fetch 替代。模块脚本等规则也与经典脚本不同，不能概括所有 script 都不受跨源约束。

开发代理和反向代理让浏览器面对同源入口，后端再转发；它们没有“关闭浏览器同源策略”。跨窗消息可以用 postMessage，接收方需校验 origin 和来源对象，发送方使用明确 targetOrigin。

localStorage 按源划分，现代隐私策略还可能进一步分区；Cookie 的域路径与站点规则不同，不能混成一套存储规则。

依据：[MDN：Same-origin policy](https://developer.mozilla.org/en-US/docs/Web/Security/Same-origin_policy)、[MDN：CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/CORS)。
