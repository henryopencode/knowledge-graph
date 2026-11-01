---
title: "前端面试 P066：自定义指令与 DOM 资源清理"
kind: "concept"
tags: ["前端面试", "Vue", "自定义指令", "DOM"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "46af9459-afbc-49b7-aaf9-299702a0d73a"
---

来源：[方格前端讲面试 · 24.Vue的自定义指令](https://www.bilibili.com/video/BV14Ttw6WEkt?p=66)，8 分 57 秒。

回看定位：[00:49 注册](https://www.bilibili.com/video/BV14Ttw6WEkt?p=66&t=49) · [01:44 钩子](https://www.bilibili.com/video/BV14Ttw6WEkt?p=66&t=104) · [07:47 清理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=66&t=467)

## 适用于可复用的底层 DOM 行为

自动聚焦、第三方原生交互接入等可以封装成指令。纯数据逻辑通常用组合式函数；有结构和交互界面的功能通常用组件；本来用 :style 就能清楚表达的内容也不必强行变指令。

```vue
<script setup>
const vFocus = { mounted: element => element.focus() };
</script>
<template><input v-focus aria-label="搜索"></template>
```

script setup 中 vFocus 这样的顶层命名可在模板使用；应用级可以 app.directive('focus', definition)。Options API 局部注册的选项是 directives，不是单数 directive。

## 钩子与参数

常见 mounted、updated、unmounted 分别处理初始化、值或组件更新后的调整、卸载清理。binding 包含 value、oldValue、arg、modifiers 等；元素行为围绕 el 进行。

updated 不代表仅当某个值严格变化才调用。若初始化昂贵的第三方实例，应避免每次 updated 重建，按需要比较配置并保存实例，卸载时销毁。

## 边界与清理

指令注册的监听、观察器、定时器应在对应销毁阶段解除。组件上使用自定义指令有根节点与多根限制，优先将指令施加到清楚的真实元素上。

v-permission 隐藏或禁用按钮只影响界面，不是安全授权；服务端接口必须独立鉴权。SSR 中需要考虑服务端渲染行为，不能在顶层无条件调用 DOM API。

本集复用 DOM 行为与清理建议正确；这里补充局部选项名称、使用范围和界面权限边界。

依据：[Vue：自定义指令](https://vuejs.org/guide/reusability/custom-directives.html)。
