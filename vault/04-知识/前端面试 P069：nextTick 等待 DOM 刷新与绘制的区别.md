---
title: "前端面试 P069：nextTick 等待 DOM 刷新与绘制的区别"
kind: "concept"
tags: ["前端面试", "Vue", "nextTick", "DOM", "调度"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "e4befa8d-2871-4f54-8eff-1c182bbfe9e1"
---

来源：[方格前端讲面试 · 27.nextTick的作用是什么？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=69)，6 分 13 秒。

回看定位：[01:42 DOM 需求](https://www.bilibili.com/video/BV14Ttw6WEkt?p=69&t=102) · [03:06 状态的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=69&t=186) · [03:55 批处理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=69&t=235)

## 数据先改变，组件 DOM 后刷新

响应式赋值后，数据本身通常可立刻读到新值；组件 DOM 更新由调度器合并。nextTick 用于等待相关的 DOM 更新刷新，之后再读取新节点或执行聚焦等动作。

```js
import { ref, nextTick } from 'vue';
const count = ref(0);
count.value = 1;
console.log(count.value); // 1，数据不需要等 nextTick
await nextTick();
// 有组件依赖该值时，可在这里读取其刷新后的 DOM。
```

先修改数据，再 await nextTick；在没有已排队更新前调用，不能保证等待所有未来尚未发生的操作。

## 需要纠正的三个概括

nextTick 不是让响应式状态赋值本身“真正生效”，数据已变；也不是它负责把多次变更合并，合并主要由框架调度完成；更不保证浏览器已经把画面绘制到屏幕。

因此把 loading 设 true、await nextTick 后立刻做巨大同步计算，不保证用户一定看见加载提示，仍可能在绘制前阻塞线程。

## 不等待什么

nextTick 不自动等待图片下载、字体就绪、CSS 动画、异步组件或网络请求完成。布局测量、视觉帧、资源加载各有对应生命周期和 API，需要按真实条件等待。

watch 中需要访问更新后 DOM，也可考虑 flush: 'post'。不要为每次赋值都追加 nextTick，避免把正常数据流拆成难维护的异步链。

依据：[Vue：nextTick](https://vuejs.org/api/general.html#nexttick)。关联：[[前端面试 P021：事件循环、微任务与渲染机会]]、[[前端面试 P050：Vue 生命周期、DOM 时机与 SSR]]。
