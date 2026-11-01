---
title: "前端面试 P089：Redux middleware、next 与异步返回值"
kind: "concept"
tags: ["前端面试", "React", "Redux", "middleware", "thunk"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "b1ab3b79-e882-4e9f-ad03-f1e1b928f882"
---

来源：[方格前端讲面试 · 19.Redux 的中间件（Middleware）机制是如何工作的？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=89)，12 分 18 秒。

回看定位：[04:03 三层](https://www.bilibili.com/video/BV14Ttw6WEkt?p=89&t=243) · [07:29 thunk](https://www.bilibili.com/video/BV14Ttw6WEkt?p=89&t=449) · [10:01 组合](https://www.bilibili.com/video/BV14Ttw6WEkt?p=89&t=601)

## 三层函数与调用链

middleware 接收包含 getState、dispatch 的 API，返回接收 next 的函数，再返回处理 action 的函数。next 把当前输入交给后续链，dispatch 则通常从增强链入口重新派发。

```js
const logger = ({ getState }) => next => action => {
  const before = getState();
  const result = next(action);
  const after = getState();
  console.log({ before, action, after });
  return result;
};

const thunk = ({ dispatch, getState }) => next => action =>
  typeof action === 'function' ? action(dispatch, getState) : next(action);
```

这个 thunk 片段说明基本机制，真实库还可注入额外依赖等。函数 action 被执行后，可以在恰当时机再 dispatch 请求开始、成功或失败的普通对象。

## 纠正“next 会等待所有后续工作”

普通 next 调用只是函数调用。若后续返回 Promise，不写 await 的代码不会自动等待请求结束；after 可能只代表此次同步 dispatch 后状态，不是整条异步业务最终状态。

保留返回值很重要，调用方可能需要等待 thunk 的 Promise。拦截不向下传递也应是明确选择，不能无意吞掉 action 或错误。

## 它不是 reducer 自带副作用许可

副作用放在 middleware、事件或其他业务层，reducer 继续纯计算。收到的 API 也不一定是包含所有方法的完整 store 实例，主要契约是 dispatch/getState。

middleware 通常按注册顺序进入，返回路径反向退出；组合函数从右向左构造，不代表最右中间件先收到普通派发。现代 Toolkit 的默认 middleware 与配置也要保留，别为添加 logger 意外移除 thunk。

依据：[Redux：middleware](https://redux.js.org/understanding/history-and-design/middleware)。
