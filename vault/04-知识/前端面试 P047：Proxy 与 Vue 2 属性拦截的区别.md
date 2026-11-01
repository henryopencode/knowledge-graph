---
title: "前端面试 P047：Proxy 与 Vue 2 属性拦截的区别"
kind: "concept"
tags: ["前端面试", "Vue", "Proxy", "Object.defineProperty"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "28274038-a160-463e-a762-6e8682b92fb4"
---

来源：[方格前端讲面试 · 5.Vue响应式底层原理-Proxy](https://www.bilibili.com/video/BV14Ttw6WEkt?p=47)，10 分 17 秒。

回看定位：[03:01 拦截概念](https://www.bilibili.com/video/BV14Ttw6WEkt?p=47&t=181) · [04:18 defineProperty](https://www.bilibili.com/video/BV14Ttw6WEkt?p=47&t=258) · [05:12 数组](https://www.bilibili.com/video/BV14Ttw6WEkt?p=47&t=312) · [07:10 依赖追踪](https://www.bilibili.com/video/BV14Ttw6WEkt?p=47&t=430)

## Proxy 拦截经过代理的操作

```js
const raw = { count: 0 };
const reads = [];
const proxy = new Proxy(raw, {
  get(target, key, receiver) {
    reads.push(key);
    return Reflect.get(target, key, receiver);
  },
  set(target, key, value, receiver) {
    return Reflect.set(target, key, value, receiver);
  },
});
proxy.count++;
console.log(raw.count); // 1
console.log(reads.includes('count')); // true
```

直接修改 raw 不会经过 proxy 的 trap；Proxy 也不会自动深度代理所有嵌套对象。Vue reactive 在此基础上管理代理缓存、嵌套访问、集合与依赖追踪等行为。

## defineProperty 的能力边界

Object.defineProperty 可以为已有属性定义 getter/setter，包括数组的某些索引；但它不是对整个对象的任意新增、删除操作设置统一拦截。

因此视频说它“原生监听不到数组下标”不够准确：语言能力并非完全不支持已有索引访问器，Vue 2 选择的数组响应式策略与全对象 Proxy 是不同的设计。

Vue 2 增强的主要变异方法包含 push、pop、shift、unshift、splice、sort、reverse；slice 本身不改原数组，不能误记成相同的变异方法。

## 检测变化不等于自动更新界面

trap 只提供观察和控制操作的入口。还需要知道哪个计算或组件曾读取属性、怎样记录依赖、怎样调度与清理。直接在 set 里固定更新某个 DOM 的小例子，不是 Vue 全部实现。

Reflect.get/set 的 receiver 有助于保留访问器中的接收者语义；set 应返回正确布尔结果，而不是无条件声称写入成功。

关联：[[前端面试 P048：track、trigger、动态依赖与清理]]、[[前端面试 P029：对象原型、prototype 与属性查找]]。
