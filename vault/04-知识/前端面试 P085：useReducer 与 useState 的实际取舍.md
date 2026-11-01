---
title: "前端面试 P085：useReducer 与 useState 的实际取舍"
kind: "concept"
tags: ["前端面试", "React", "useReducer", "reducer"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "96183ff8-bf97-4a08-9576-d350efb5735f"
---

来源：[方格前端讲面试 · 15. useReducer 相比useState优势体现在哪？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=85)，14 分 00 秒。

回看定位：[05:08 转换函数](https://www.bilibili.com/video/BV14Ttw6WEkt?p=85&t=308) · [08:23 测试](https://www.bilibili.com/video/BV14Ttw6WEkt?p=85&t=503) · [10:26 稳定性的原比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=85&t=626)

## reducer 集中定义状态转换

把相关数据与转换规则集中，UI 负责 dispatch 意图，纯 reducer 计算下一份状态，便于独立测试与复用。它更适合规则多、转换有关联的状态，不是 useState 的全面升级。

```js
function reducer(state, action) {
  switch (action.type) {
    case 'increment': return { ...state, count: state.count + state.step };
    case 'decrement': return { ...state, count: Math.max(0, state.count - state.step) };
    case 'reset': return { ...state, count: 0 };
    default: return state;
  }
}
console.log(reducer({ count: 2, step: 3 }, { type: 'increment' }).count); // 5
console.log(reducer({ count: 2, step: 3 }, { type: 'decrement' }).count); // 0
```

## 必须纠正：useState 的 setter 也稳定

视频把性能区别归结为 useState 的 setState 每次变化、useReducer 的 dispatch 才稳定，这不正确。两者返回的更新函数身份通常都稳定；新建的 handleIncrement 包装器可能变化，是另一回事。

dispatch 稳定不代表子组件一定不渲染，其他 props、context 和自身状态仍会影响。选择 reducer 的主要收益是规则与可维护性，不是自动获得更少渲染。

## 纯度与结构

reducer 不发请求、不修改旧 state。异步工作在事件、Effect 或相应层处理后 dispatch 结果，返回旧对象可表示无变化。复杂规则可以独立函数化，useState 也能使用纯计算函数并测试，不是无法获得同类可测试性。

useState 与 reducer 的内部关联属于实现细节，不能把某个版本如何复用代码当公共 API 保证。

依据：[React：useReducer](https://react.dev/reference/react/useReducer)、[React：useState](https://react.dev/reference/react/useState)。
