---
title: "前端面试 P058：路由参数、query 与组件 props"
kind: "concept"
tags: ["前端面试", "Vue", "VueRouter", "路由参数"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "b569ce18-e35a-4c53-ae2d-463eec3e5c42"
---

来源：[方格前端讲面试 · 16.Vue-Router的路由参数与传参方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=58)，4 分 14 秒。

回看定位：[00:13 三种方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=58&t=13) · [00:36 动态参数](https://www.bilibili.com/video/BV14Ttw6WEkt?p=58&t=36) · [00:21 query](https://www.bilibili.com/video/BV14Ttw6WEkt?p=58&t=21)

## params 与 query 是 URL 的不同部分

`/users/:id` 的 id 通常标识资源；`/search?q=vue&page=2` 的 query 常表示搜索、筛选和分页。params 可以通过 ?、*、+ 等配置成可选或重复，并非所有动态参数都必填。

```js
import { createRouter, createMemoryHistory } from 'vue-router';
const router = createRouter({ history: createMemoryHistory(), routes: [
  { path: '/users/:id', name: 'user', component: {}, props: true },
] });
await router.push({ name: 'user', params: { id: 42 }, query: { tab: 'posts' } });
console.log(router.currentRoute.value.params.id); // '42'
console.log(router.currentRoute.value.query.tab); // posts
```

URL 中的类型要解析和验证，不能把 query.page 自动当作可靠数字。对象形式传 path 时，params 通常不会替你拼接；参数导航优先用 route name 或自己构造有效路径。

## 路由 props 不是第三套独立 URL 编码

props: true 把 params 映射成组件 props；函数模式可将 params/query 转换成组件需要的输入，减少组件直接依赖 useRoute。它不是用来把任意复杂对象偷偷存入 URL 的机制。

同一动态记录的参数改变时，组件可能被复用，不能只在 mounted 读取一次；监听相关参数、使用路由更新守卫或派生 props，处理新旧请求竞态。

不要在 URL 放密钥或敏感业务原文，它可能进入历史、日志、截图与分享链接。复杂临时数据可放相应状态或明确持久化，但刷新恢复语义要另行设计。

依据：[Vue Router：路由组件传参](https://router.vuejs.org/guide/essentials/passing-props.html)。关联：[[前端面试 P059：导航守卫、顺序与权限边界]]。
