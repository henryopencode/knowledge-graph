---
title: "前端面试 P062：Pinia 创建、实例范围与使用"
kind: "concept"
tags: ["前端面试", "Vue", "Pinia", "store"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "507dea40-9d8d-4c8e-9ce5-8c7d5670efac"
---

来源：[方格前端讲面试 · 20.Pinia基本用法-store](https://www.bilibili.com/video/BV14Ttw6WEkt?p=62)，5 分 04 秒。

回看定位：[01:04 defineStore](https://www.bilibili.com/video/BV14Ttw6WEkt?p=62&t=64) · [02:48 获取实例](https://www.bilibili.com/video/BV14Ttw6WEkt?p=62&t=168) · [03:48 唯一范围](https://www.bilibili.com/video/BV14Ttw6WEkt?p=62&t=228)

## 定义 store，再在 Pinia 范围内取得实例

应用通常先 createPinia 并 app.use(pinia)。defineStore 返回 useStore 函数；同一个 Pinia 容器内，使用相同 store 定义通常得到同一实例。

```js
import { createPinia, defineStore } from 'pinia';
const pinia = createPinia();
const useCounter = defineStore('counter-demo', {
  state: () => ({ count: 0 }),
  actions: { increment() { this.count++; } },
});
const first = useCounter(pinia);
const second = useCounter(pinia);
first.increment();
console.log(second.count); // 1
console.log(first === second); // true
```

组件外或测试中可明确传入 pinia，避免没有活动容器的调用时机问题。store ID 应在相应容器内唯一，不能让两个不同 store 意外争用同一个 ID。

## “全局唯一”需要限定

课程说不管哪里调用都拿到全局唯一实例，应限定为同一 Pinia 容器；不同应用或不同 SSR 请求应有各自容器。模块级立即调用 useStore 还可能早于插件安装。

## Options store 与 Setup store

Options 形式用 state、getters、actions；Setup store 可返回 ref、computed 和函数来表达相应角色。两者都要明确导出的状态和行为，SSR、插件、重置支持也有具体差别。

store 的值可以在模板中读取，直接 store.count 不需要外加 .value。需要解构状态时使用 storeToRefs，见 [[前端面试 P064：storeToRefs 与 action 解构纠错]]。

依据：[Pinia：定义 Store](https://pinia.vuejs.org/core-concepts/)。
