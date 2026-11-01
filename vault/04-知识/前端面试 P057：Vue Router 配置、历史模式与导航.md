---
title: "前端面试 P057：Vue Router 配置、历史模式与导航"
kind: "concept"
tags: ["前端面试", "Vue", "VueRouter", "路由"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "21c656f0-be78-4b14-bd9a-795217d80153"
---

来源：[方格前端讲面试 · 15.Vue-Router的基本配置与使用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=57)，4 分 43 秒。

回看定位：[01:15 配置](https://www.bilibili.com/video/BV14Ttw6WEkt?p=57&t=75) · [02:35 导航方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=57&t=155)

## URL 与组件匹配

Vue Router 4 使用 createRouter、history 与 routes 建立匹配关系；应用 app.use(router) 后，RouterView 是匹配组件出口，RouterLink 提供声明式导航。

```js
import { createRouter, createMemoryHistory } from 'vue-router';
const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/', name: 'home', component: { template: '<p>Home</p>' } },
    { path: '/about', name: 'about', component: { template: '<p>About</p>' } },
  ],
});
await router.push('/about');
console.log(router.currentRoute.value.name); // about
```

memory history 便于测试或相应服务端环境；浏览器常用 createWebHistory 或 createWebHashHistory。history 路径需要服务器正确回退应用入口，避免直接刷新深层路径时 404。

## 声明式与编程式

RouterLink 适合可见导航，router.push 适合逻辑触发；replace 替换当前历史记录，go/back/forward 操作历史。普通外链仍使用正确的 a 链接，不应把所有跳转都改成按钮调用。

导航可能异步、取消、重定向或失败，重要流程应等待 router.push 的返回结果，而不是调用后立即假定已到达目标。滚动行为、404 路由和懒加载也应根据应用配置。

本集基础方向正确，但“单页切换不刷新”不代表硬刷新、外部导航或所有资源请求都消失。

关联：[[前端面试 P058：路由参数、query 与组件 props]]。
