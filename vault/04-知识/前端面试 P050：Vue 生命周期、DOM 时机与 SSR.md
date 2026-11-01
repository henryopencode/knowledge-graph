---
title: "前端面试 P050：Vue 生命周期、DOM 时机与 SSR"
kind: "concept"
tags: ["前端面试", "Vue", "生命周期", "SSR"]
created: "2026-10-26T09:57:12Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "d30b4051-aa59-48f7-8cdb-8b755e3b7c84"
---

来源：[方格前端讲面试 · 8.Vue常用的生命周期函数](https://www.bilibili.com/video/BV14Ttw6WEkt?p=50)，10 分 39 秒。

回看定位：[02:09 生命周期阶段](https://www.bilibili.com/video/BV14Ttw6WEkt?p=50&t=129) · [03:54 清理配对](https://www.bilibili.com/video/BV14Ttw6WEkt?p=50&t=234) · [06:26 父子顺序](https://www.bilibili.com/video/BV14Ttw6WEkt?p=50&t=386) · [09:44 SSR 原建议](https://www.bilibili.com/video/BV14Ttw6WEkt?p=50&t=584)

## 常用三对钩子

| 阶段 | 更新前后 | 典型用途 |
| --- | --- | --- |
| 首次挂载 | onBeforeMount、onMounted | 首次渲染前准备；挂载后访问相应 DOM、初始化客户端库 |
| 组件更新 | onBeforeUpdate、onUpdated | 读取旧 DOM；更新完成后检查 DOM |
| 组件卸载 | onBeforeUnmount、onUnmounted | 释放监听、连接、定时器等资源 |

setup 负责组件初始化逻辑，不等于 onMounted。生命周期钩子通常应在有效的同步 setup 上下文注册，不能随意等一个异步回调以后再注册。

## 初始化与清理配对

```vue
<script setup>
import { ref, onMounted, onUnmounted } from 'vue';
const width = ref(0);
const resize = () => { width.value = window.innerWidth; };
onMounted(() => {
  resize();
  window.addEventListener('resize', resize);
});
onUnmounted(() => window.removeEventListener('resize', resize));
</script>
<template><p>{{ width }}</p></template>
```

保存同一个 resize 引用，才能正确解绑。可在 onBeforeUnmount 需要 DOM 仍可用时提前清理，具体由资源语义决定。

## 纠正 SSR 的浏览器 API 建议

视频结尾先正确说 onMounted 不在服务端执行，随后又说浏览器 API 因此不能放这里，这个推论反了。onMounted 恰好是常见的客户端专用访问位置；真正要避免的是在会被服务端执行的模块顶层或 setup 路径无条件访问 window、document。

服务端首屏数据也不能只依靠 onMounted；需要结合框架的数据获取、onServerPrefetch 等方案。初始化内容还应注意客户端与服务端水合一致性。

## 父子顺序与更新限制

普通同步子树通常父先 beforeMount、子先 mounted；异步组件、Suspense、SSR 等需要另行分析，父 mounted 不保证所有异步后代都已完成。

onUpdated 对应组件 DOM 更新之后，不要与 beforeUpdate 混淆。为了等待某次特定状态变化后的 DOM，nextTick 或合适的 watcher flush 通常比在所有 updated 回调里做业务更明确。不要在 updated 中无条件改渲染依赖造成循环。

依据：[Vue：生命周期钩子](https://vuejs.org/api/composition-api-lifecycle.html)。关联：[[前端面试 P051：副作用清理与过期请求]]。
