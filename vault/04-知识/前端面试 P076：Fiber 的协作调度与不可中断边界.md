---
title: "前端面试 P076：Fiber 的协作调度与不可中断边界"
kind: "concept"
tags: ["前端面试", "React", "Fiber", "调度", "并发渲染"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "1c6ebfa1-725d-440c-bb1a-6642b62b6ffb"
---

来源：[方格前端讲面试 · 6.React 的 Fiber 架构主要是为了解决什么问题？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=76)，10 分 57 秒。

回看定位：[01:11 旧同步机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=76&t=71) · [04:00 拆分](https://www.bilibili.com/video/BV14Ttw6WEkt?p=76&t=240) · [05:24 两个阶段](https://www.bilibili.com/video/BV14Ttw6WEkt?p=76&t=324) · [08:31 优先级](https://www.bilibili.com/video/BV14Ttw6WEkt?p=76&t=511)

## 为什么引入 Fiber

老式递归同步协调难以在庞大更新中让出执行机会。Fiber 以工作单元和可维护的执行状态组织协调，为优先级、可恢复或重启的渲染与并发特性提供基础。

## render 与 commit

render 阶段计算候选树与变化，相关并发工作可被暂停、重新开始或丢弃；commit 对已决定的变更同步提交，不能像 render 那样随意中断并展示半套结果。各阶段的 Effect、布局读取也有具体时序。

## 需要补充的限制

不是所有更新都自动异步分片，也不是每执行一个 Fiber 必然让浏览器休息。优先级和更新来源影响调度，transition 等用来表达部分更新可不紧急。

React 不能在一个正在执行的任意用户函数内部随时抢占。组件函数里一个巨大同步循环仍会阻塞；大量 DOM 提交同样可能耗时。Fiber 不是多线程执行 JS，也不会把计算自动搬进 Worker。

调度器不等同 requestIdleCallback，每个版本和运行目标的实现还会演进；“只在浏览器空闲时执行”不是完整描述。

## 纯度为什么关键

既然候选渲染可以被重做或放弃，render 里发请求、改外部对象等行为就可能重复或无法回滚。保持渲染纯粹，外部同步放在正确边界，才能安全配合调度。

关联：[[前端面试 P071：UI 是状态映射与渲染纯度]]、[[前端面试 P081：StrictMode 的检查与版本边界]]。
