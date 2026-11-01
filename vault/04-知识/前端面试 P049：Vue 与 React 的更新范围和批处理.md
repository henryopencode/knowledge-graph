---
title: "前端面试 P049：Vue 与 React 的更新范围和批处理"
kind: "concept"
tags: ["前端面试", "Vue", "React", "批处理", "渲染"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "f164aae3-c4b5-4e95-b14a-435d3d209c78"
---

来源：[方格前端讲面试 · 7.Vue 响应式更新vsReact 状态更新](https://www.bilibili.com/video/BV14Ttw6WEkt?p=49)，14 分 13 秒。

回看定位：[01:30 更新范围](https://www.bilibili.com/video/BV14Ttw6WEkt?p=49&t=90) · [03:31 Vue 调度](https://www.bilibili.com/video/BV14Ttw6WEkt?p=49&t=211) · [05:01 React 批处理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=49&t=301) · [10:50 常见陷阱](https://www.bilibili.com/video/BV14Ttw6WEkt?p=49&t=650)

## 比较触发、计算与提交三件事

Vue 常通过响应式读写关联依赖，React 常通过 setter 提交状态更新。组件函数或 render 重新执行，不等于所有对应 DOM 节点都被重建；最后还有比较、调度和提交。

Vue 数据通常可同步读到新值，组件 DOM 更新常合并到调度刷新中；nextTick 可等待相关刷新。React 当前回调中的 state 是当前渲染快照，setter 请求下一次更新，不会原地改变这份变量。

## 批处理不等于所有赋值累加

React 同一处理器里多次 setCount(count + 1) 使用同一个旧快照，通常最终只加一；setCount(value => value + 1) 多次排队才逐步基于前项结果累加。

Vue 连续 count.value++ 则直接修改响应式当前值，DOM 刷新仍可合并。不要只比较控制台打印次数，就把状态计算语义也当成相同。

## 更新范围取决于数据和组件结构

课程演示 Vue 只更新第三项、React 父更新带动全部子组件，这有对应结构前提。Vue 父模板读取位置、传递的 props、对象替换方式等会影响范围；React 也可利用稳定子元素、memo、状态下移或编译优化等减少工作。

React.memo 还可能因自身状态或 context 更新重新执行；同样内容但新创建的对象、函数会影响浅比较。不能把“Vue 永远只更新一个”“React 永远全部更新”当作框架定律。

## flushSync 与开发统计

React 现代根支持更广泛的自动批处理，但不保证跨所有异步边界合成一次。flushSync 用于少数需要同步提交的集成场景，不宜为了固定渲染次数滥用；StrictMode 和并发渲染也会影响开发日志。

Vue 3.5 的响应式 props 解构是编译特例；视频对所有 props 解构都丢响应性的概括，应注明版本与写法。

关联：[[前端面试 P046：Vue 响应式陷阱与 ref 自动解包纠错]]、[[前端面试 P023：async、await、异步恢复与并发]]。
