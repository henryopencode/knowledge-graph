---
title: "前端面试 P025：Promise 解析过程与实现验收"
kind: "concept"
tags: ["前端面试", "JavaScript", "Promise", "thenable", "边界"]
created: "2026-10-26T09:01:31Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "98aa1a7f-3619-5fa5-8316-530aae81ec24"
---

来源：[方格前端讲面试 · 25. 手写一个简化版 Promise](https://www.bilibili.com/video/BV14Ttw6WEkt?p=25)，24 分 36 秒。

## 为什么 resolve 不能永远直接保存 value

原生 Promise 必须吸收 Promise/thenable 的结果，而不是把它们当普通值永久保存。课程为了简化而省略这一步，需要明确说明后果。

```js
Promise.resolve(1)
  .then(value => Promise.resolve(value + 1))
  .then(console.log); // 2，而不是把内部 Promise 当作最终值交出
```

基础状态机 [[前端面试 P025：手写 Promise 的最小状态机]] 不具备这种完整解析行为。

## 解析时要分别考虑什么

1. 当前要解析的 Promise 不能解析为自己，否则用 TypeError 拒绝。
2. 原始值可直接作为成功值。
3. 对象或函数需要安全读取 then；getter 本身可能抛错。
4. then 可调用时，以该对象作接收者调用，并使用只允许首次决定的保护。
5. then 成功回调传入的值还可能是另一个 thenable，需要继续解析。
6. 调用 then 抛错时，若先前已成功决定，后来的异常不能覆盖结果。

## 最小验收清单

检查 executor 同步、then 异步、先注册后决定与先决定后注册、多订阅者、状态只决定一次、值穿透、原因穿透、回调抛错、返回 Promise、读取 then 抛错、自身循环和重复调用。

不要只测一条“延迟后成功”的链。处理器返回值、错误路径和恶意 thenable 才是实现最容易出错的地方。

原生 Promise 还涉及 species、宿主未处理拒绝跟踪、构造器等语义，不能把 A+ 的解析过程当成全部语言规范。学习手写是理解机制，业务继续用经过验证的实现。

参考：[Promises/A+](https://promisesaplus.com/)、[ECMAScript Promise 对象](https://tc39.es/ecma262/#sec-promise-objects)。
