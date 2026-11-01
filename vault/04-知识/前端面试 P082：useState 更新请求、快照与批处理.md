---
title: "前端面试 P082：useState 更新请求、快照与批处理"
kind: "concept"
tags: ["前端面试", "React", "useState", "批处理"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "dec13952-bc14-441c-b048-dd06c589b296"
---

来源：[方格前端讲面试 · 12.useState的更新是异步的吗？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=82)，8 分 48 秒。

回看定位：[01:51 快照](https://www.bilibili.com/video/BV14Ttw6WEkt?p=82&t=111) · [00:27 批处理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=82&t=27) · [06:01 React 18](https://www.bilibili.com/video/BV14Ttw6WEkt?p=82&t=361)

## 不要只用同步或异步一个词回答

setter 调用提交更新，当前渲染的 state 变量不会原地改变。何时计算、渲染、提交由 React 调度，多个更新可以批处理；setter 本身不返回代表 DOM 完成的 Promise。

```jsx
function Counter() {
  const [count, setCount] = React.useState(0);
  return <button onClick={() => {
    setCount(count + 1);
    console.log(count); // 本次初始点击仍为 0
  }}>{count}</button>;
}
```

旧值主要由渲染快照与闭包解释，不是只要强制同步提交就能把这份局部 count 变量改掉。

## 自动批处理的范围

React 18 现代 root 扩大到 Promise、定时器、原生事件等场景的批处理；不代表任意跨事件、跨 await 的所有更新永远只有一次 render。不同交互仍需保持行为正确，开发检查与可中断渲染也可能产生额外计算。

flushSync 可用于少数必须及时提交 DOM 的外部集成，但不是应对旧闭包读值的通用修复。不要 await setCount 期待等待一次 React 提交。

## 与课程结论的区别

视频用“延迟处理而非异步”纠正简单口号有帮助，但同样不能反向保证任何更新都总被延迟。回答应分别说明调用、当前变量快照、更新排队与 DOM 提交。

依据：[React：useState](https://react.dev/reference/react/useState)。关联：[[前端面试 P074：函数组件快照与类实例的输出纠错]]、[[前端面试 P083：函数式更新、队列与纯度]]。
