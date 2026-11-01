---
title: "前端面试 P090：Zustand、Jotai 与状态作用域"
kind: "concept"
tags: ["前端面试", "React", "Zustand", "Jotai", "状态管理"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "c3f4668c-0d3c-479d-8611-310d84b06aaa"
---

来源：[方格前端讲面试 · 20.你是否了解像 Zustand、Jotai 这类新兴的状态管理库？它们各自有什么特点？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=90)，14 分 37 秒。

回看定位：[01:17 两种模型](https://www.bilibili.com/video/BV14Ttw6WEkt?p=90&t=77) · [05:08 Provider 原比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=90&t=308) · [08:25 atom](https://www.bilibili.com/video/BV14Ttw6WEkt?p=90&t=505)

## 两种常见模型

Zustand 常围绕 store、set/get 与 selector 组织共享状态；Jotai 常从 atom 与派生 atom 组织依赖。前者不是只能有一个巨大 store，后者的值也由具体 store 容器管理，不是 atom 定义对象里永远只有一份全局值。

## 需要纠正 Provider 的绝对比较

Zustand 的常用全局 store 可以不用 Provider；按子树、应用或请求隔离时，仍可结合 Context 等提供实例。Jotai 也支持 provider-less 的默认 store，Provider 主要帮助显式隔离作用域，不能说 Jotai 一般必须顶层包裹才可工作。

SSR 用户状态应按请求隔离，别因为“不需 Provider”就把服务端进程级可变状态当成默认安全方案。

## 精确订阅仍有接口条件

Zustand selector 结果变化决定相关通知，返回每次新建对象时要考虑相等策略与 useShallow 等工具；原地改深层数据、不更新引用也可能破坏预期。

Jotai 的派生 atom 可以记录依赖，但 atom 身份需稳定，不能在 render 每次无条件创建不稳定 atom 并期待自动正常。自动追踪不是任何状态图都零成本。

## Hook 与外部 API

React 中使用这些库的 Hook 仍遵守 Hooks 规则；组件外通过库提供的 store.getState、get/set 等 API 访问，不能把“任何地方可用”理解成可以在普通函数任意调用 Hook。

选择时比较数据图、消费者范围、调试、SSR 与团队维护需要，而不是按“新库默认更优”决定。

依据：[Zustand 文档](https://zustand.docs.pmnd.rs/)、[Jotai Provider](https://jotai.org/docs/core/provider)。
