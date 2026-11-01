---
title: "前端面试 P015：IIFE、作用域隔离与现代模块"
kind: "concept"
tags: ["前端面试", "JavaScript", "IIFE", "模块化"]
created: "2026-10-26T08:49:37Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "cd7b1cbf-4e22-5463-a5f6-e53c3ffbd728"
---

来源：[方格前端讲面试 · 15. 立即执行函数 IIFE 是怎么工作的？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=15)，18 分 19 秒。

回看定位：[01:05 定义](https://www.bilibili.com/video/BV14Ttw6WEkt?p=15&t=65) · [06:37 表达式与调用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=15&t=397) · [16:10 现代模块](https://www.bilibili.com/video/BV14Ttw6WEkt?p=15&t=970)

## IIFE 的结构

立即调用函数表达式把函数作为表达式求值后立即调用，常见写法是 `(function () { ... })()`。第一组括号使当前位置解析成表达式，后一组执行调用。

```js
const result = (function (value) {
  const privateValue = value * 2;
  return privateValue + 1;
})(5);
console.log(result); // 11
```

它有普通函数的参数与返回值。外部拿不到局部绑定，但返回对象、函数或注册监听器等动作仍可能把内部数据暴露出去，隔离作用域不是安全沙箱。

## 早期解决什么问题

在 var 普遍使用时，IIFE 方便创建函数作用域、避免临时变量进入全局，也可以用返回的函数实现模块封装。循环里每次把 i 作为参数传入，也能为延迟回调建立不同的参数绑定。

```js
const counter = (() => {
  let value = 0;
  return () => ++value;
})();
console.log(counter()); // 1
console.log(counter()); // 2
```

这里 IIFE 只运行一次，返回的 counter 以后运行多次；二者不能混淆。闭包原理见 [[前端面试 P011：闭包、变量绑定与循环回调]]。

## 语法和异步边界

表达式上下文不只由外层括号创建，void 等写法也可以，但优先选直观形式。与上一行可能被连成表达式时，应保留分号，避免自动分号插入误解。

async IIFE 返回 Promise，内部 await 不会让外部代码自动等待。需要处理它的 rejection；ES 模块支持顶层 await 时，也可根据模块依赖关系选择更直接的结构。

视频将函数说成早期唯一能创建作用域的结构，是面向 var 隔离的简化描述；catch 参数等历史语义也有局部绑定。现代代码通常优先使用 ES 模块与 let/const，IIFE 仍会出现在初始化表达式、打包产物和遗留代码中。
