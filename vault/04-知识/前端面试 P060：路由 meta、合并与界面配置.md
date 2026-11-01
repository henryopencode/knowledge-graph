---
title: "前端面试 P060：路由 meta、合并与界面配置"
kind: "concept"
tags: ["前端面试", "Vue", "VueRouter", "meta"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "d584bcd5-aa12-46fa-9786-387e972f679f"
---

来源：[方格前端讲面试 · 18.Vue-Router的路由元信息](https://www.bilibili.com/video/BV14Ttw6WEkt?p=60)，5 分 29 秒。

回看定位：[00:54 meta 配置](https://www.bilibili.com/video/BV14Ttw6WEkt?p=60&t=54) · [02:41 标题](https://www.bilibili.com/video/BV14Ttw6WEkt?p=60&t=161) · [03:50 敏感信息](https://www.bilibili.com/video/BV14Ttw6WEkt?p=60&t=230)

## meta 是业务约定，不是内建权限执行器

路由记录可以定义 title、requiresAuth、menuVisible 等自定义元信息。读取 route.meta 或守卫中的 to.meta，再由应用代码决定标题、导航提示、菜单或其他行为。

```js
const routes = [{
  path: '/admin', component: AdminLayout,
  meta: { requiresAuth: true, layout: { section: 'admin' } },
  children: [{ path: 'users', component: UserList, meta: { title: '用户' } }],
}];
```

这是配置片段。AdminLayout、UserList 是实际组件；只设置 requiresAuth 不会自动拦截请求，仍需要守卫和后端授权。

## 父子 meta 的合并

route.meta 通常是匹配记录从父到子的非递归合并，子层同名字段覆盖父层，不会自动深合并 layout 等嵌套对象。需要分层策略时查看 route.matched 的各记录，明确继承和覆盖规则。

标题更新应兼顾失败导航、默认值和运行环境。document.title 是客户端行为，SSR 的文档头管理要使用适合服务端渲染的方案。

## 类型与边界

为 RouteMeta 建立 TypeScript 类型，统一字段名称与默认规则。不要把密钥、权威权限结论放到前端配置；隐藏菜单不等于禁止访问，requiresAuth 也不是服务器的可信凭证。

本集自定义元信息的主线正确；补充浅合并、匹配记录和安全边界，避免把“元数据配置”误认成框架自动完成全部业务。

关联：[[前端面试 P059：导航守卫、顺序与权限边界]]。
