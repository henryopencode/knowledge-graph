---
title: "前端面试 P080：组合、children 与自定义 Hook"
kind: "concept"
tags: ["前端面试", "React", "组合", "自定义Hook"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "dc1207b9-7d3a-4a35-90dc-ae79c89630fe"
---

来源：[方格前端讲面试 · 10.为什么在React中“组合优于继承”](https://www.bilibili.com/video/BV14Ttw6WEkt?p=80)，12 分 26 秒。

回看定位：[05:19 children](https://www.bilibili.com/video/BV14Ttw6WEkt?p=80&t=319) · [06:35 多区域](https://www.bilibili.com/video/BV14Ttw6WEkt?p=80&t=395) · [00:45 自定义 Hook](https://www.bilibili.com/video/BV14Ttw6WEkt?p=80&t=45)

## UI 用组件接口组合

children 可以表达容器内容，具名 prop 可以传 header、footer、操作区域等元素。容器不必继承每种具体页面，只要提供清楚的结构与行为接口。

```jsx
function Card({ title, children, actions }) {
  return <section><h2>{title}</h2><div>{children}</div><footer>{actions}</footer></section>;
}
```

这与类组件 extends React.Component 的基础使用不是矛盾；“组合优于继承”主要针对用业务组件继承层级复用 UI 的设计方式。

## 逻辑用 Hook 组合，但状态不自动共享

自定义 Hook 可以封装订阅、状态与清理，每个调用通常拥有独立的 Hooks 状态。两个组件都调用 useOnlineStatus，不代表它们因为函数名相同就共享同一份 useState。

需要真正共享的外部数据，可使用 context、外部 store 或底层订阅源；Hook 只是访问逻辑的复用入口。订阅的建立、解除、参数变化都要设计。

## 组合也有契约成本

插槽和回调并不保证零耦合。提供过多含糊 prop、层层 Hook 依赖，仍会让理解困难。遵守 Hooks 规则，在组件或自定义 Hook 合适的顶层调用，不能因它是普通函数语法就任意在条件、循环中调用。

本集推荐组合的方向正确；共享逻辑、共享状态和共享外部资源应分开说明。

依据：[React：用自定义 Hook 复用逻辑](https://react.dev/learn/reusing-logic-with-custom-hooks)。关联：[[前端面试 P044：Composition API 与组合式函数]]。
