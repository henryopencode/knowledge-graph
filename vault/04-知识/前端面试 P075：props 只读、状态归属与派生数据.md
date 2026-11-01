---
title: "前端面试 P075：props 只读、状态归属与派生数据"
kind: "concept"
tags: ["前端面试", "React", "props", "不可变更新"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "4d729649-af27-4763-b9f1-cd1d62661862"
---

来源：[方格前端讲面试 · 5. 为什么React中props不可变](https://www.bilibili.com/video/BV14Ttw6WEkt?p=75)，12 分 32 秒。

回看定位：[02:38 单向数据](https://www.bilibili.com/video/BV14Ttw6WEkt?p=75&t=158) · [09:23 初始 state 的原建议](https://www.bilibili.com/video/BV14Ttw6WEkt?p=75&t=563)

## props 是当前渲染收到的输入

子组件应把 props 当只读快照，需要变更时调用明确的回调，请状态拥有者更新。修改嵌套对象会影响共享引用，但不会自动通知 React 更新，更不会保证别的组件立刻出现“幽灵重渲染”。

## 不可变更新支持比较，但不是默认深比较

视频说可变 props 会迫使 React 深比较，且引用没变就百分百跳过整个组件树，这不准确。React 不会因你原地改对象就自动替你深比较；普通父渲染仍可能调用子组件，memo/PureComponent 等优化有具体比较规则，还会受自身状态、context 等影响。

按修改路径创建新对象，使新旧快照可区分，是清晰数据流和相关优化的重要前提。

## 不要一概把 props 复制进 state

视频推荐“要修改时永远先用 props 初始化内部 state”，过于绝对。useState(initialProp) 通常只用初始值初始化，之后 prop 改变不会自动同步。

可直接派生的数据尽量在 render 计算；由父控制的输入使用 value/onChange；确实需要独立编辑草稿时，才明确初始化、保存、取消与切换对象后的重置规则。

```jsx
function NameInput({ value, onChange }) {
  return <input value={value} onChange={event => onChange(event.target.value)} />;
}
```

这个受控接口不需要复制一份相同状态。需要为另一个用户重置编辑草稿时，也可根据正确身份使用 key，而不是盲目在 Effect 中同步所有 props。

关联：[[前端面试 P077：key、组件身份与状态重置]]、[[前端面试 P079：受控表单与非受控输入]]。
