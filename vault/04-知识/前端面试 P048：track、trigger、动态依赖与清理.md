---
title: "前端面试 P048：track、trigger、动态依赖与清理"
kind: "concept"
tags: ["前端面试", "Vue", "响应式", "依赖追踪", "effect"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "463da526-23c6-48fd-810b-4e39516534a5"
---

来源：[方格前端讲面试 · 6.Vue响应式底层原理-依赖追踪机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=48)，11 分 39 秒。

回看定位：[02:36 track 条件](https://www.bilibili.com/video/BV14Ttw6WEkt?p=48&t=156) · [06:27 依赖结构](https://www.bilibili.com/video/BV14Ttw6WEkt?p=48&t=387) · [07:23 effect](https://www.bilibili.com/video/BV14Ttw6WEkt?p=48&t=443) · [11:10 生命周期](https://www.bilibili.com/video/BV14Ttw6WEkt?p=48&t=670)

## 依赖怎样产生

某个 effect 活跃执行时读取响应式属性，系统把该 effect 与目标、属性关联，这一步可称 track。属性改变后找到相关订阅者，通知重新计算或进入调度器，可称 trigger。

课程用 WeakMap<target, Map<key, Set<effect>>> 解释关联关系，适合理解索引结构；这是教学模型，不应当据此断言每个 Vue 版本内部都逐字使用相同数据结构。

## 分支会改变依赖

```js
import { reactive, watchEffect, nextTick } from 'vue';
const state = reactive({ show: true, left: 1, right: 2 });
const seen = [];
const stop = watchEffect(() => {
  seen.push(state.show ? state.left : state.right);
});
state.show = false;
await nextTick();
state.left = 9;
await nextTick();
console.log(seen); // [1, 2]
stop();
```

切到 right 后，left 不再是这次计算真正需要的依赖；如果只永久往集合 add、不清理旧依赖，就会不断多触发。

## 基础手写模型还缺什么

- 嵌套 effect 需要保存与恢复上层活跃状态，而不是执行完直接清空全部关联。
- effect 抛错时也要用 finally 恢复追踪上下文。
- 遍历依赖集合同时清理再添加，可能导致重复执行，需要稳定的迭代与调度策略。
- 相同值、数组 length、遍历键、删除、Map/Set 等需要专门处理。
- 要有停止与作用域销毁机制，并避免无界自触发。

因此“get 登记、set 直接执行全部函数”只是起点。Vue 组件更新还会合并与去重，不保证每次赋值同步重跑所有 DOM 工作。

watchEffect 的异步函数通常只自动追踪同步执行、首次 await 前读取的依赖；不能把 await 后的任意读取也当作自动登记。

关联：[[前端面试 P047：Proxy 与 Vue 2 属性拦截的区别]]。
