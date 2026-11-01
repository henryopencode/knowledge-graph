---
title: "前端面试 P063：Pinia 状态、getter 与 action"
kind: "concept"
tags: ["前端面试", "Vue", "Pinia", "state", "action"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "ce1ccf95-cbe5-4511-a39c-3da124d84744"
---

来源：[方格前端讲面试 · 21.Pinia三大核心概念](https://www.bilibili.com/video/BV14Ttw6WEkt?p=63)，6 分 01 秒。

回看定位：[00:50 state](https://www.bilibili.com/video/BV14Ttw6WEkt?p=63&t=50) · [02:09 getter](https://www.bilibili.com/video/BV14Ttw6WEkt?p=63&t=129) · [05:40 限制的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=63&t=340)

## 三类成员职责

state 保存可变业务数据；getter 从已有响应式数据派生结果，通常应保持纯；action 表达业务操作，可执行异步任务，也可调用其他 store 或 action。

```js
import { createPinia, defineStore } from 'pinia';
const useCounter = defineStore('core-demo', {
  state: () => ({ count: 0 }),
  getters: { double: state => state.count * 2 },
  actions: { increment() { this.count++; } },
});
const store = useCounter(createPinia());
store.count = 2;
console.log(store.double); // 4
store.$patch({ count: 3 });
console.log(store.double); // 6
store.increment();
console.log(store.count); // 4
```

## 必须纠正：不是只有 action 才能改状态

视频结尾说“只有 actions 里面才可以修改 state”，这不符合 Pinia 的 API。直接修改状态与 $patch 都是支持的；把业务变更集中进 action 是团队约定和可追踪性的选择，不是库强制禁止其他写法。

getter 的缓存依赖实际响应式读取；返回一个接受参数的函数时，函数内部的每个参数结果不自动全部被记忆化。需要参数缓存时另行设计。

## 更新分组与异步

$patch 可以表达分组修改，尤其函数形式修改多个集合操作；不能无条件声称必定比直接赋值性能更高。async action 返回 Promise，调用方仍需等待或处理错误；store 不自动处理请求乱序、取消和权限。

Options action 需要 this 时用普通方法写法，箭头函数没有所需自己的 this。对外 API、错误与状态重置也应明确。

依据：[Pinia：State](https://pinia.vuejs.org/core-concepts/state.html)、[Pinia：Actions](https://pinia.vuejs.org/core-concepts/actions.html)。
