---
title: "前端面试 P054：provide、inject 与跨层级依赖"
kind: "concept"
tags: ["前端面试", "Vue", "provide", "inject", "组件通信"]
created: "2026-10-26T09:57:12Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "a1940337-b7e5-4a53-a479-d93a490a2d4a"
---

来源：[方格前端讲面试 · 12.Provide-Inject的用法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=54)，9 分 02 秒。

回看定位：[01:26 跨层传值](https://www.bilibili.com/video/BV14Ttw6WEkt?p=54&t=86) · [05:19 主题示例](https://www.bilibili.com/video/BV14Ttw6WEkt?p=54&t=319) · [07:15 使用场景](https://www.bilibili.com/video/BV14Ttw6WEkt?p=54&t=435)

## 谁提供，谁能注入

祖先在有效组件上下文 provide 一个 key 对应的值，后代 inject 同一个 key；最近匹配的祖先会遮蔽更远的同名提供者。应用级 app.provide 还能提供默认依赖。

它是组件树范围的依赖注入，不是任何组件随时互相发现的全局事件总线。兄弟组件需要共同祖先提供，或选择其他状态机制。

## 响应式由提供的值决定

提供 ref 或 reactive 对象可保留相应响应式关系；提供 ref.value 读出的普通字符串，通常只是当时值。inject 得到 ref 时，脚本中仍按 ref 的 .value 访问，不会因为注入就自动变普通值。

```js
// 在提供者的 setup 中；themeKey 从共享 key 模块导入
import { ref, readonly, provide } from 'vue';
const theme = ref('light');
provide(themeKey, {
  theme: readonly(theme),
  toggle: () => { theme.value = theme.value === 'light' ? 'dark' : 'light'; },
});
```

后代 inject(themeKey) 获取同一个接口，通过显式 toggle 请求修改。key 可用 Symbol 减少冲突，但两边必须使用同一个 Symbol 实例，不能各自临时 Symbol('theme')。

## 默认值、缺失与生命周期

没有提供者时应决定是否允许默认值，复杂默认对象可以使用工厂形式。TypeScript 下 InjectionKey 可表达类型，不会自动处理运行时缺失。

把修改逻辑留在提供者，通常比让任意后代直接修改共享对象更易追踪。需要订阅、持久化、DevTools、跨页面业务状态等能力时，Pinia 等工具可能更合适。

示例中的“所有注入组件自动更新”还需要它们实际读取相应响应式数据；提供一个对象并不意味着每次变动都强制所有后代重渲染。SSR 也要避免把每请求私有状态放进进程共享单例。

依据：[Vue：依赖注入](https://vuejs.org/guide/components/provide-inject.html)。关联：[[前端面试 P044：Composition API 与组合式函数]]。
