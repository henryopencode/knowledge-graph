---
title: "前端面试 P043：Vue 渐进式与声明式 UI"
kind: "concept"
tags: ["前端面试", "Vue", "Vue3", "声明式UI"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "9a06b0ab-354b-4afe-9ff6-c9567fe92885"
---

来源：[方格前端讲面试 · 1.Vue 的设计哲学 & 核心理念00](https://www.bilibili.com/video/BV14Ttw6WEkt?p=43)，7 分 09 秒。

回看定位：[00:14 渐进式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=43&t=14) · [03:18 性能](https://www.bilibili.com/video/BV14Ttw6WEkt?p=43&t=198) · [04:25 响应式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=43&t=265)

## 渐进式表示按需要采用

可以为已有页面引入局部交互，也可以结合单文件组件、路由、状态管理与构建工具组织完整应用。不必从第一天就加入所有生态工具，也不意味着小项目永远不需要工程化。

## 声明式界面

```vue
<script setup>
import { ref } from 'vue';
const count = ref(0);
</script>

<template>
  <button @click="count++">{{ count }}</button>
  <p v-if="count >= 10">达到十次</p>
</template>
```

开发者描述状态、事件和显示条件，框架负责把变化应用到 DOM。脚本中的 ref 读取常用 count.value，模板在相应位置自动解包；这不是普通 JavaScript 变量改变就能自动改 DOM。

## 响应式与更新并非同一层

Vue 3 对对象主要使用 Proxy，ref 的 value 访问也有相应追踪机制。数据变化通知依赖，调度器再合并组件更新；并非每次赋值都立即完成 DOM 修改。

静态提升、补丁标记等编译优化能减少工作，但不能只凭框架名字保证任意业务都更快。仍要关注组件设计、长列表、数据规模和真实测量。

## 与 React 的比较应限定层面

Vue 常以响应式读取追踪依赖，React 常通过状态 setter 提交更新。两者都支持声明式 UI、组件复用和批处理，不宜把差别说成“一个自动、另一个所有事情都手动”。Vue 的 watch 与 React 的 Effect 也不是完全等价。

视频的渐进式与声明式介绍适合作为入门；这里补充响应式和调度层次，避免把性能和更新粒度的概括当作无条件保证。

依据：[Vue：简介](https://vuejs.org/guide/introduction.html)。关联：[[前端面试 P049：Vue 与 React 的更新范围和批处理]]。
