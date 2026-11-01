---
title: "前端面试 P041：事件委托、目标匹配与 React 事件"
kind: "concept"
tags: ["前端面试", "浏览器", "事件委托", "React", "DOM"]
created: "2026-10-26T09:13:34Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "44ce864a-779d-5754-8e63-8f7f1d15863e"
---

来源：[方格前端讲面试 · 41.事件委托的本质(完结）](https://www.bilibili.com/video/BV14Ttw6WEkt?p=41)，8 分 03 秒。

回看定位：[01:23 传播阶段](https://www.bilibili.com/video/BV14Ttw6WEkt?p=41&t=83) · [02:30 target](https://www.bilibili.com/video/BV14Ttw6WEkt?p=41&t=150) · [05:30 React 17](https://www.bilibili.com/video/BV14Ttw6WEkt?p=41&t=330) · [06:08 Portal](https://www.bilibili.com/video/BV14Ttw6WEkt?p=41&t=368)

## 委托利用传播，但不应只看 target.dataset

把处理器放到稳定父节点，由事件传播到父节点时识别实际操作。event.target 是原始目标（也可能因 Shadow DOM 被重定向），currentTarget 是当前注册监听器的节点。

```js
const list = document.querySelector('#list');
list.addEventListener('click', event => {
  if (!(event.target instanceof Element)) return;
  const button = event.target.closest('button[data-id]');
  if (!button || !list.contains(button)) return;
  console.log(button.dataset.id);
});
```

点击按钮内部的图标时，target 可能是图标，不一定带 data-id。closest 和范围检查避免漏识别或匹配到列表外的祖先。实际业务还要按操作类型验证载荷。

## 哪些事件适合

动态列表中的冒泡事件很适合委托，新增子节点通常不需逐个绑定。focus/blur、mouseenter 等传播规则不同，可根据目标改用 focusin/focusout、捕获阶段或其他相应事件，不能机械套用 click 写法。

stopPropagation、嵌套委托、Shadow DOM 的 composed 与 retargeting 也会影响结果。监听器数减少不等于整页性能瓶颈必然解决，渲染和列表规模可能更重要。

## 纠正视频对 React 列表的性能推导

React 常见事件已通过框架委托，JSX map 中为每项写 onClick，不等于真的在每个 DOM 节点各挂一份同类原生监听。可能仍有回调分配和组件更新成本，但不能直接用“1000 个 JSX onClick 就是 1000 个原生监听器”作为论据。

React 17 起，多数委托事件的根位置由 document 调整到根容器；不是所有事件都完全统一采用一种注册方式，也不保证多个应用从此没有任何交互。

## Portal 有两条传播关系

React Portal 的合成事件可按 React 组件树传播；底层原生事件仍有物理 DOM 路径。不能据此说原生事件不再经过 body 或物理祖先。混用原生与 React 监听时，需要分别分析。

依据：[React：createPortal](https://react.dev/reference/react-dom/createPortal)、[React 17 事件委托变更](https://legacy.reactjs.org/blog/2020/10/20/react-v17.html#changes-to-event-delegation)。
