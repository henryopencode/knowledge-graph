---
title: "前端面试 P087：Redux 单向数据流与状态边界"
kind: "concept"
tags: ["前端面试", "React", "Redux", "不可变更新"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "8a5bb398-2b48-4f97-9f0f-8b046f224082"
---

来源：[方格前端讲面试 · 17.说一下Redux的核心思想](https://www.bilibili.com/video/BV14Ttw6WEkt?p=87)，10 分 37 秒。

回看定位：[02:23 三原则](https://www.bilibili.com/video/BV14Ttw6WEkt?p=87&t=143) · [05:15 action](https://www.bilibili.com/video/BV14Ttw6WEkt?p=87&t=315) · [08:01 数据流](https://www.bilibili.com/video/BV14Ttw6WEkt?p=87&t=481)

## 三条核心原则

集中式 store 提供需要共享的状态来源；通过 dispatch action 描述变化意图；纯 reducer 根据当前 state 和 action 计算下一份状态。不要在 reducer 里请求、修改旧对象或依赖不可控时间随机数。

```js
function reducer(state = { name: 'Alice' }, action) {
  if (action.type === 'user/nameChanged') {
    return { ...state, name: action.payload };
  }
  return state;
}
const previous = reducer(undefined, { type: 'init' });
const next = reducer(previous, { type: 'user/nameChanged', payload: 'Bob' });
console.log(previous.name); // Alice
console.log(next.name); // Bob
```

## 不是应用里每份状态都要进 Redux

视频“整个应用所有状态都放一个 store”是三原则的常见简化，不代表局部输入、临时交互必须入库。是否共享、是否需要追踪与持久化，应按状态归属判断。

action 通常是带字符串 type 的普通对象，middleware 可以接收其他输入并最终派发普通 action。纯 reducer 不必每次创建新对象，无变化时应保留旧引用，修改路径外的数据也可结构共享。

## Redux 与 React 更新不是同一件事

store 接受派发后更新订阅者，React-Redux 再通过 selector 与相应比较决定组件是否需要更新；不能把每次 dispatch 等同全部组件重渲染。可追溯、时间旅行与持久化也依赖序列化、工具与业务设计，不会因单个 store 自动全部完成。

现代 Redux 通常优先使用 Redux Toolkit，简化 reducer 与开发检查；Immer 相关说明见 [[前端面试 P088：Immer 草稿与结构共享]]。
