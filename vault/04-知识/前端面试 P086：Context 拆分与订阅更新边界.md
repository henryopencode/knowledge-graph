---
title: "前端面试 P086：Context 拆分与订阅更新边界"
kind: "concept"
tags: ["前端面试", "React", "Context", "性能"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "abac1f47-2e73-4e1d-ad6d-73881f4814a7"
---

来源：[方格前端讲面试 · 16.如何优化因 useContext 引起的性能问题？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=86)，10 分 30 秒。

回看定位：[00:57 广播](https://www.bilibili.com/video/BV14Ttw6WEkt?p=86&t=57) · [03:46 拆分](https://www.bilibili.com/video/BV14Ttw6WEkt?p=86&t=226) · [07:57 分离 API](https://www.bilibili.com/video/BV14Ttw6WEkt?p=86&t=477)

## Context 比较提供的 value

Provider 的 value 用 Object.is 等相应语义比较；新对象身份会让读取该 context 的消费者看到更新。useContext 本身不为任意对象字段自动建立 selector 式细分订阅。

## 三种策略

按职责和变化频率拆分 context；用 useMemo 等稳定确有必要的提供值；将状态与稳定的 dispatch/action 接口分开，让只负责操作的组件不订阅整份 state。

组合 children 还可隔离由某个状态拥有者引起的普通父子渲染，但前提是传入元素身份在该次更新中稳定。

## 不能由这些策略推出什么

memo 不会挡住组件实际订阅的 context 更新。稳定 children 也不会阻止后代自己的 state 或 context 变化。状态与 dispatch 分开，只是减少来自这份 context 的通知，不是按钮永远不再执行。

useMemo 是性能优化手段，不是应用正确性依赖的永久缓存。不要遗漏依赖只为维持引用，导致消费者看到过期数据。

## 更细订阅

高频、字段多、消费者分散时可考虑 selector 型外部 store 或相应方案。首先用 Profiler 确认耗时位置，再决定拆分，避免把几十份 context 变成另一种维护负担。

本集策略可用；这里主要纠正“隔离后永远不会再更新”的绝对保证。

关联：[[前端面试 P080：组合、children 与自定义 Hook]]、[[前端面试 P085：useReducer 与 useState 的实际取舍]]。
