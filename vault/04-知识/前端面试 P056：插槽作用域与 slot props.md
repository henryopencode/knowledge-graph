---
title: "前端面试 P056：插槽作用域与 slot props"
kind: "concept"
tags: ["前端面试", "Vue", "slot", "作用域插槽"]
created: "2026-10-26T09:57:12Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "6950c817-596b-4439-918d-7f1ec36709fb"
---

来源：[方格前端讲面试 · 14.插槽作用域](https://www.bilibili.com/video/BV14Ttw6WEkt?p=56)，7 分 15 秒。

回看定位：[00:41 作用域原则](https://www.bilibili.com/video/BV14Ttw6WEkt?p=56&t=41) · [04:01 slot props](https://www.bilibili.com/video/BV14Ttw6WEkt?p=56&t=241) · [05:02 数据归属](https://www.bilibili.com/video/BV14Ttw6WEkt?p=56&t=302)

## 渲染位置不改变词法归属

父模板声明的插槽内容使用父模板可见的绑定，不会自动获得子组件内部变量。子组件愿意公开哪些数据，就通过 slot 出口的属性传出，使用者用 v-slot 参数接收。

可以类比为：父提供一个渲染函数，子调用它时传入 slot props；函数仍能访问创建它的外部绑定，同时也能使用参数。

## 子组件提供列表行

```vue
<script setup>
defineProps({ rows: { type: Array, default: () => [] } });
</script>
<template>
  <ul>
    <li v-for="(row, index) in rows" :key="row.id">
      <slot :row="row" :index="index">{{ row.name }}</slot>
    </li>
  </ul>
</template>
```

父组件使用 `<DataList :rows="users">`，内部的 `<template #default="{ row, index }">` 可读取 row、index，也仍可读取父组件自己的 prefix 等变量。具名插槽同样可以带参数。

## 不是任意读取子内部状态

没有作为 slot props 暴露的数据不会自动进入父插槽作用域。slot 的 name 是选择出口的特殊属性，不会作为普通同名 slot prop 直接传给使用方。

暴露函数也要定义它改变什么状态、是否能被任意重复调用。把完整可变内部对象交出去会增加耦合，按所需最小接口提供更易维护。

## 纠正异常表述的边界

视频把访问未提供的子变量概括为“报错”。具体表现可能是 Vue 警告、读取到 undefined 或最终表达式触发异常，取决于模板表达式；不能把“没显示”直接当作已经抛出了同一种错误。

本集关于词法归属与显式传参的主要方向正确，见 [[前端面试 P009：执行上下文、调用栈与词法作用域]]、[[前端面试 P055：默认插槽、具名插槽与内容接口]]。
