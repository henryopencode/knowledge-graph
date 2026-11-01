---
title: "前端面试 P095：AI SDK、编排框架与生成 UI 工具"
kind: "concept"
tags: ["前端面试", "前端工程", "AI", "SDK", "工具选型"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "bc80079f-dc99-462e-8502-9667bc4ee0e6"
---

来源：[方格前端讲面试 · 你了解哪些AI前端工具或库](https://www.bilibili.com/video/BV14Ttw6WEkt?p=95)，6 分 11 秒。

回看定位：[00:43 SDK 分类](https://www.bilibili.com/video/BV14Ttw6WEkt?p=95&t=43) · [03:39 编排](https://www.bilibili.com/video/BV14Ttw6WEkt?p=95&t=219) · [05:26 生成工具](https://www.bilibili.com/video/BV14Ttw6WEkt?p=95&t=326)

## 先按职责区分

| 类型 | 视频例子 | 主要处理什么 |
| --- | --- | --- |
| 提供商官方 SDK | OpenAI JavaScript/TypeScript SDK 等 | 访问相应提供商 API、类型与请求能力 |
| 应用 SDK | Vercel AI SDK | 模型接入、流式处理与聊天界面等集成能力 |
| 编排与检索框架 | LangChain.js 等 | 工具、工作流、检索与应用编排 |
| 生成式开发产品 | v0 | 根据描述或参考生成界面与代码 |

v0 是产品与开发工具，不应直接当作普通可导入 UI 组件库。官方 SDK、Agent 编排、UI Hooks 与设计生成也不是同一层功能。

## 官方资料核对

OpenAI 官方文档把 SDK 用于应用 API 请求，Agents SDK 用于相应 agent 编排；JavaScript SDK 安装包是 openai，示例在 Node.js、Deno、Bun 等服务端环境使用。

这不意味着应把服务端长期 API key 打包进浏览器。通常由自己的后端保存密钥，处理鉴权、额度、工具权限与流式转发；浏览器展示状态、输入、消息和取消操作。文档见 [OpenAI SDKs and CLI](https://developers.openai.com/api/docs/libraries)。

## 纠正过时名称和混合边界

视频提到的模型、图片能力与包 API 名称会随版本变化，不能把旧例子当作 2026 固定清单。AI SDK 的 UI 与核心包有不同职责，LangChain 的编排也不自动保证检索正确或工具调用安全。

选型时确认具体版本、浏览器/服务端运行环境、流协议、工具能力、供应商替换成本、可观测性和测试方式，不以库名数量代替实际项目理解。

## 面试如何讲实战

说明你使用的工具在调用链哪一层，解决了什么重复工作，如何处理超时、错误、取消、上下文与成本；生成的代码如何编译、测试和审查。不要只展示“输入一句话就生成页面”。

其他参考：[AI SDK 文档](https://ai-sdk.dev/docs/introduction)、[LangChain JavaScript](https://docs.langchain.com/oss/javascript/langchain/overview)。关联：[[前端面试 P093：MCP 架构与前端集成场景]]。
