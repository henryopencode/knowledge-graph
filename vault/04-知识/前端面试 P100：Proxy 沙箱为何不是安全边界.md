---
title: "前端面试 P100：Proxy 沙箱为何不是安全边界"
kind: "concept"
tags: ["前端面试", "前端工程", "微前端", "Proxy", "隔离"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "7f51b2fc-b3da-4969-8513-10a1823b09da"
---

来源：[方格前端讲面试 · 微前端架构中，一般是如何到JavaScript隔离的？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=100)，12 分 46 秒。

## 拦截属性访问不等于复制整个环境

```js
const host = { shared: { value: 1 } };
const local = Object.create(null);
const scope = new Proxy(local, {
  get: (target, key) => Object.hasOwn(target, key) ? target[key] : host[key],
  set: (target, key, value) => Reflect.set(target, key, value),
});
scope.name = 'app A';
console.log(host.name); // undefined
scope.shared.value = 2;
console.log(host.shared.value); // 2
```

第一种直接属性写入被放进 local；第二种先读到 host.shared，再改其内部对象，没有经过顶层 set，仍会影响共享环境。这与 [[前端面试 P003：原始值、对象共享与赋值行为]] 的对象共享规则一致。

## 哪些能力仍可能共享

document、事件目标、原型对象、存储、网络凭据以及真实全局引用都可能使代码产生外部影响。把函数的 this 绑定到 proxy，也不能让它内部所有词法与全局引用自动重定向。

使用 eval/new Function 等执行脚本还会引入 CSP、严格模式、作用域与可信代码来源等问题；不能仅展示几条可拦截赋值，就声称能安全运行任意不可信 JavaScript。

## 教学模拟与生产框架

课程的代理与快照模式适合理解隔离意图，不是完整沙箱实现。框架还需处理生命周期、异步回调包装、原生方法绑定、资源回收和多个应用兼容性。

真正的敌对代码隔离，要依赖适当安全边界与权限设计，不能把同 realm 内的便利代理当成浏览器安全模型的替代。

关联：[[前端面试 P100：微前端 JS 隔离方案与资源生命周期]]、[[前端面试 P039：同源策略、CORS 与请求是否发出]]。
