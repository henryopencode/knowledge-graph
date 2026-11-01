---
title: "前端面试 P081：StrictMode 的检查与版本边界"
kind: "concept"
tags: ["前端面试", "React", "StrictMode", "Effect", "清理"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "025d7906-73f8-4aef-8ac1-32a03cec2572"
---

来源：[方格前端讲面试 · 11.React 的严格模式 (StrictMode) 有什么作用？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=81)，9 分 50 秒。

回看定位：[01:32 开发检查](https://www.bilibili.com/video/BV14Ttw6WEkt?p=81&t=92) · [04:46 重复调用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=81&t=286) · [05:43 Effect 清理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=81&t=343) · [06:20 旧 API](https://www.bilibili.com/video/BV14Ttw6WEkt?p=81&t=380)

## 开发检查，不生成额外 DOM

StrictMode 开启开发阶段额外检查，包括重复调用部分应纯粹的逻辑、额外 Effect 清理重建、ref 回调检查与相关弃用提示。生产不执行这些开发检查。

```jsx
createRoot(document.getElementById('root')).render(
  <React.StrictMode><App /></React.StrictMode>,
);
```

render 不能依赖“只被调用一次”保证，事件处理器也不是所有情况下都被 StrictMode 自动调用两次。不同 React 版本、root 或子树包裹位置可能影响额外检查，不能固定背成任何逻辑都准确两遍。

## Effect 为何 setup → cleanup → setup

额外重建用于检查外部订阅和资源清理是否对称，帮助暴露重复连接等问题。关闭 StrictMode 隐藏日志，不会修好遗漏清理。

这不是生产环境必然真卸载整个 DOM 再重新挂载的证明，也不是用 useRef 强行阻止第二次 setup 就完成兼容。

## 视频中旧 API 要按版本记

findDOMNode、旧 context、字符串 ref 等在老版本可能是警告或弃用，在 React 19 有些已移除，不能只当成“当前仍可用但 StrictMode 会提示”。React 19 的函数组件 ref 传递方式也有变化，不必把 forwardRef 说成所有现代场景唯一方案。

## 不能提供的保证

StrictMode 不会自动找出所有副作用或证明线程安全。应实际测试清理、重入、取消和错误路径。正确代码能承受重复渲染与作用域重新建立，而不是只调到日志看起来一次。

依据：[React：StrictMode](https://react.dev/reference/react/StrictMode)。关联：[[前端面试 P076：Fiber 的协作调度与不可中断边界]]。
