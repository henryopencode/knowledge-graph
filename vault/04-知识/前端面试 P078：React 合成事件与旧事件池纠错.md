---
title: "前端面试 P078：React 合成事件与旧事件池纠错"
kind: "concept"
tags: ["前端面试", "React", "合成事件", "事件委托"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "98573e92-9c8b-4ea0-894f-eff2b9a9b239"
---

来源：[方格前端讲面试 · 8.React 的事件机制和合成事件是如何工作的？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=78)，13 分 29 秒。

回看定位：[03:49 委托](https://www.bilibili.com/video/BV14Ttw6WEkt?p=78&t=229) · [06:25 事件池的旧说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=78&t=385) · [08:19 persist](https://www.bilibili.com/video/BV14Ttw6WEkt?p=78&t=499)

## 合成事件做什么

React 提供较一致的事件 API，并根据内部树和宿主事件组织处理器。多数委托事件在现代 React DOM 根容器处理，但有捕获、非委托事件等差异，不是全应用只有一个原生监听函数。

## 必须纠正：现代 React DOM 没有旧式事件池

视频同时讲 React 17 根容器委托与旧事件池，版本混用了。React 17 起，React DOM 已取消旧式事件池，通常不需 e.persist 才能异步保留合成事件；该方法在这里不再起原来的作用。

旧版本池化时存在事件回调后属性被清空的问题，但不能把它作为“2026 新面试”现代 React 的默认行为。

```jsx
function Button() {
  function onClick(event) {
    const type = event.type;
    setTimeout(() => console.log(type, event.type), 0);
  }
  return <button onClick={onClick}>查看事件</button>;
}
```

保存所需原始值仍是清楚的做法。注意 currentTarget 与某些 DOM 引用的生命周期、节点是否还在，不要把“没有池化”误当作事件所有字段永远不变。

## 原生传播与 React 传播分开

preventDefault 处理默认行为，stopPropagation 处理相应传播；它们不是一回事。原生监听顺序、根容器、多 root 与 Portal 等都需结合真实位置分析，不是简单“两个冒泡必然互不影响”。

JSX 列表每项 onClick 不等于每项各添加一份同类原生监听，见 [[前端面试 P041：事件委托、目标匹配与 React 事件]]。

依据：[React DOM：事件对象](https://react.dev/reference/react-dom/components/common#react-event-object)、[旧事件池说明](https://legacy.reactjs.org/docs/legacy-event-pooling.html)。
