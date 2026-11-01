---
title: "前端面试 P094：流式文本的 SSE 与 WebSocket 取舍"
kind: "concept"
tags: ["前端面试", "前端工程", "SSE", "WebSocket", "流式响应"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "5964ab04-2f6f-4035-be8c-1dec830e6aa9"
---

来源：[方格前端讲面试 · ChatGPT 的对话功能实现，为什么选择 SSE协议而非Websocket](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94)，10 分 59 秒。

回看定位：[01:04 通信模式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94&t=64) · [05:28 协议区别](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94&t=328) · [06:46 重连](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94&t=406) · [07:12 资源比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94&t=432)

## 先确认通信需求

提问后持续接收生成文本，通常可以用 HTTP 请求与流式响应；持续双向、高频交换消息时，WebSocket 提供合适的双向通道。选择还取决于认证、代理、断线恢复、载荷和后端架构。

| 维度 | SSE/HTTP 事件流 | WebSocket |
| --- | --- | --- |
| 主要数据方向 | 当前响应中服务端持续下发 | 建连后双向消息 |
| 内容格式 | text/event-stream 事件格式 | 文本或二进制消息 |
| 浏览器入口 | EventSource，或 fetch 读取相应事件流 | WebSocket API |
| 重连恢复 | 原生 EventSource 有相应机制；fetch 需应用实现 | 通常由应用定义重连与恢复 |
| 应用语义 | 仍需定义事件名、结束、错误、恢复 | 同样需定义消息协议与状态 |

## 必须分清 EventSource 与 fetch POST

原生 EventSource 通常以 GET 建立连接，不能随意添加请求体与任意请求头。聊天应用可以使用 fetch POST 并读取 SSE 格式响应，但那不是“EventSource 自动帮 POST 重连”。

断线后自动重试也不等于生成可以安全续写；要设计事件 ID、会话、幂等和恢复位置，避免重复内容或重复计费操作。

## 需要修正视频的保证

不能只凭单向/双向就断言 SSE 的 CPU、内存永远更少；两者都可能维持长连接，开销需实测。HTTP/1 连接限制与 HTTP/2 多路复用也不同，反向代理的缓冲和超时会影响流式体验。

WebSocket 不是所有语音、图片或双向场景的唯一方案，HTTP、WebRTC 等也各有用途。具体产品不同功能和版本可能采用不同传输，课程标题不能作为 ChatGPT 全部功能始终使用某一种协议的证据。

## 前端必须处理的细节

网络 chunk 不等于完整事件、完整 JSON 或一个模型 token。正确解码字节、缓存半行、按事件边界解析，再更新 UI；支持主动取消、错误展示和合理渲染节流。

参考：[MDN：SSE](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)。实现示意见 [[前端面试 P094：SSE 分片解码与事件边界]]。
