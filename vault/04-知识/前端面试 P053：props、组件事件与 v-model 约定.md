---
title: "前端面试 P053：props、组件事件与 v-model 约定"
kind: "concept"
tags: ["前端面试", "Vue", "props", "emit", "v-model"]
created: "2026-10-26T09:57:12Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "870aa6e8-68fa-475b-906f-1517efe895bb"
---

来源：[方格前端讲面试 · 11.Vue的通信基础](https://www.bilibili.com/video/BV14Ttw6WEkt?p=53)，12 分 31 秒。

回看定位：[03:59 props](https://www.bilibili.com/video/BV14Ttw6WEkt?p=53&t=239) · [06:36 事件的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=53&t=396) · [07:54 v-model](https://www.bilibili.com/video/BV14Ttw6WEkt?p=53&t=474)

## 数据归属与更新请求

父组件通过 props 提供值，子组件通过 emit 表达事件或更新请求，父组件决定如何修改状态。props 是浅只读入口，不会把传入对象的全部内部数据冻结；技术上能改嵌套对象，也不意味着符合清楚的数据归属。

## 纠正：组件事件不是 DOM 冒泡

视频把 emit 事件说成冒泡到父组件，容易误导。Vue 自定义组件事件不会自动穿过整棵组件树冒泡；通常由直接使用该组件的监听者接收，需要跨层时显式转发、使用注入或共享状态。

emit 可传多个参数，定义 emits/defineEmits 有助于表达组件契约与验证。事件声明不是身份权限验证，也不宜只解释成“为了性能”。

## Vue 3 的组件 v-model

默认约定是 modelValue prop 与 update:modelValue 事件；Vue 2 常见的 value/input 不能不加区分照搬。视频本段口述和课件把 input/value 与组件写法混在一起，下面明确采用 Vue 3 约定。

```vue
<script setup>
const props = defineProps({ modelValue: String });
const emit = defineEmits(['update:modelValue']);
</script>
<template>
  <input :value="props.modelValue"
    @input="emit('update:modelValue', $event.target.value)">
</template>
```

父组件用 `<SearchInput v-model="query" />`，等价于相应的 prop 绑定与更新事件监听。具名 v-model:query 对应 query 与 update:query，也能按约定支持多个模型。

Vue 3.4+ 的 defineModel 可以简化这种接口，但仍有相应传值与更新事件语义；它不是放弃单向数据流而任意修改父状态。

## 父传值的快照问题

把 props 的原始值一次性复制到本地 ref，得到的是初始化快照，不会自动持续同步；若需要持续派生可用 computed、getter 或明确同步策略。Vue 3.5 的 defineProps 解构转换又是特定编译语义，见 [[前端面试 P046：Vue 响应式陷阱与 ref 自动解包纠错]]。

依据：[Vue：组件事件](https://vuejs.org/guide/components/events.html)、[Vue：组件 v-model](https://vuejs.org/guide/components/v-model.html)。
