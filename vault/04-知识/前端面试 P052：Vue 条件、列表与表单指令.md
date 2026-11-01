---
title: "前端面试 P052：Vue 条件、列表与表单指令"
kind: "concept"
tags: ["前端面试", "Vue", "模板", "v-if", "v-for", "v-model"]
created: "2026-10-26T09:57:12Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "d1e5fb13-7ef5-4e15-a4ef-75049c549ad2"
---

来源：[方格前端讲面试 · 10.模板语法-v-if-v-for-v-bind-v-model](https://www.bilibili.com/video/BV14Ttw6WEkt?p=52)，10 分 11 秒。

回看定位：[00:42 条件指令](https://www.bilibili.com/video/BV14Ttw6WEkt?p=52&t=42) · [02:46 稳定 key](https://www.bilibili.com/video/BV14Ttw6WEkt?p=52&t=166) · [05:27 表单](https://www.bilibili.com/video/BV14Ttw6WEkt?p=52&t=327)

## v-if 与 v-show

v-if 为真时才建立相应分支，变假会卸载其中组件；v-show 通常保留已渲染节点并切换 display。频繁显示隐藏可考虑 v-show，初始很可能不需要且创建昂贵的分支可考虑 v-if，但实际成本取决于子树。

视频“初始条件频繁改变更适合 v-if”的措辞不准确，应区分初始渲染成本与后续切换成本。

## v-for 的身份由稳定 key 表达

列表插入、删除、重排时，稳定业务 ID 有助于让输入状态和组件实例跟随正确条目。索引 key 在会重排且有内部状态的列表容易错位；静态、纯展示等场景并非语言层面禁止索引，避免把建议记成绝对语法禁令。

Vue 3 中，同元素上 v-if 的优先级高于 v-for，因此 v-if 不应直接依赖该 v-for 才产生的 item。可以预先计算过滤列表，或把结构拆到不同层。

```vue
<script setup>
import { computed, ref } from 'vue';
const rows = ref([{ id: 1, name: 'Alice', active: true }]);
const activeRows = computed(() => rows.value.filter(row => row.active));
</script>
<template>
  <ul><li v-for="row in activeRows" :key="row.id">{{ row.name }}</li></ul>
</template>
```

## v-bind 与 v-model

冒号是 v-bind 的简写，动态绑定 DOM 属性、class、style 等。绑定布尔属性要理解值本身与静态字符串的差别，不能靠把 false 拼成字符串控制全部行为。

原生表单的 v-model 根据控件类型选择 value/checked 与 input/change 等组合，也有 lazy、trim、number 等修饰符。组件 v-model 是约定的 prop 与更新事件，不是任意 input 事件都自动生效。

组件用法见 [[前端面试 P053：props、组件事件与 v-model 约定]]。条件隐藏不等于暂停所有副作用，v-show 保留组件时要额外考虑运行中的任务。

依据：[Vue：条件渲染](https://vuejs.org/guide/essentials/conditional.html)、[Vue：列表渲染](https://vuejs.org/guide/essentials/list.html)。
