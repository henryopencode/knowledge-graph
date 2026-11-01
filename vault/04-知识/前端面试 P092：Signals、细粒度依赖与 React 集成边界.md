---
title: "前端面试 P092：Signals、细粒度依赖与 React 集成边界"
kind: "concept"
tags: ["前端面试", "React", "Signals", "响应式"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "c1030429-7657-4829-819f-bf571200851e"
---

来源：[方格前端讲面试 · 22.你是如何理解 Signals 的？它与 React 现有的状态管理方式有何不同？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=92)，11 分 46 秒。

回看定位：[04:05 电子表格](https://www.bilibili.com/video/BV14Ttw6WEkt?p=92&t=245) · [03:50 push pull 原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=92&t=230) · [08:48 直接 DOM 的原结论](https://www.bilibili.com/video/BV14Ttw6WEkt?p=92&t=528) · [09:29 性能绝对说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=92&t=569)

## Signal 通常表达可追踪的值

读操作可建立依赖，写操作通知相关计算，派生值根据依赖更新。电子表格中一个单元格改变、相关公式重算，是理解细粒度响应式的好类比。

但 Signals 不是单一库或 React 内建的一个统一 API。value 属性、getter/setter、computed、effect、订阅粒度和批处理规则随实现而不同。

## 需要纠正 push/pull 的二分法

视频把 React 统一叫 push、Signals 统一叫纯 pull，过度简化。许多信号系统会推送失效、在读取时拉取新值，采用混合策略；有依赖图不等于只剩某一种更新方向。

## React 中不保证组件函数永不重跑

是否能直接更新文本节点，要看库的渲染集成、是否把 signal 对象直接作为支持的子节点、是否在组件里读取 .value 等条件。不能把示例里某条日志只打印一次，推广成“引入 Signals 后 React 一定绕过全部组件与虚拟 DOM”。

Solid、Preact、React 适配等不同方案的执行模型也不完全相同。外部 store、并发一致性、SSR 与资源清理仍需相应支持。

## 性能和开发体验的取舍

细粒度追踪可减少无关计算，但也有依赖管理、订阅与调度成本，不能承诺“默认性能最优”。动态分支、异步 effect、循环依赖、共享作用域与调试都可能带来新边界。

React 的组件更新也不等于顶层一次 setter 自动重算整棵应用，编译优化和结构设计同样影响范围。先定位瓶颈，再比较合适的状态模型。

关联：[[前端面试 P048：track、trigger、动态依赖与清理]]、[[前端面试 P049：Vue 与 React 的更新范围和批处理]]。
