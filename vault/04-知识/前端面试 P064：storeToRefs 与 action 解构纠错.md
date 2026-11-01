---
title: "前端面试 P064：storeToRefs 与 action 解构纠错"
kind: "concept"
tags: ["前端面试", "Vue", "Pinia", "storeToRefs", "解构"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "5c1ababb-e799-4bbf-a08e-2665b7411347"
---

来源：[方格前端讲面试 · 22.Pinia-Store的响应式解构](https://www.bilibili.com/video/BV14Ttw6WEkt?p=64)，7 分 27 秒。

回看定位：[01:03 状态解构](https://www.bilibili.com/video/BV14Ttw6WEkt?p=64&t=63) · [02:51 storeToRefs](https://www.bilibili.com/video/BV14Ttw6WEkt?p=64&t=171) · [05:39 action 的错误建议](https://www.bilibili.com/video/BV14Ttw6WEkt?p=64&t=339)

## 状态解构与动作解构不同

直接从 store 取出数字或字符串，通常只得到当时值；storeToRefs 可把 state、getter 等响应式成员变成保留关联的 ref。action 则已经由 Pinia 绑定，允许直接解构调用。

```js
import { createPinia, defineStore, storeToRefs } from 'pinia';
const useCounter = defineStore('destructure-demo', {
  state: () => ({ count: 0 }),
  getters: { double: state => state.count * 2 },
  actions: { increment() { this.count++; } },
});
const store = useCounter(createPinia());
const { count: snapshot } = store;
const { count, double } = storeToRefs(store);
const { increment } = store;
increment();
console.log(snapshot); // 0
console.log(count.value); // 1
console.log(double.value); // 2
```

## 必须纠正视频结论

课件和口述反复要求“永远不要解构 Pinia action，否则 this 丢失”。这是把普通对象方法的规律套到了 Pinia 包装后的动作上；对上面这种 action，官方支持解构后调用。

storeToRefs 会忽略方法及非响应式成员，因此从它返回结果里找 increment 不合适；从 store 本身取 action 即可。对象属性解构、getter 与 action 要分别判断，不能统一禁止。

## 仍需要注意什么

computed getter 对应的 ref 通常不能任意赋值。把返回的 ref.value 再一次解构成普通变量，又会形成快照。只有实际 Pinia 绑定的 action 适用这里的保证，自己额外挂上的任意普通函数仍需按其调用规则分析。

依据：[Pinia：从 store 解构](https://pinia.vuejs.org/core-concepts/#destructuring-from-a-store)。关联：[[前端面试 P010：普通函数、箭头函数与 this]]。
