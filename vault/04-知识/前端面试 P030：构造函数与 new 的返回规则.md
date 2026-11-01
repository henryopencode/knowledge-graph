---
title: "前端面试 P030：构造函数与 new 的返回规则"
kind: "concept"
tags: ["前端面试", "JavaScript", "new", "构造函数"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "23a83e7a-c51d-5d27-91e7-93823a617b9e"
---

来源：[方格前端讲面试 · 30. 构造函数与 new 的机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=30)，19 分 25 秒。

回看定位：[03:59 构造器约定](https://www.bilibili.com/video/BV14Ttw6WEkt?p=30&t=239) · [07:33 new 过程](https://www.bilibili.com/video/BV14Ttw6WEkt?p=30&t=453) · [10:59 返回值](https://www.bilibili.com/video/BV14Ttw6WEkt?p=30&t=659) · [15:31 忘记 new](https://www.bilibili.com/video/BV14Ttw6WEkt?p=30&t=931)

## 普通常规构造函数的核心过程

用 new 调用可构造目标时，为实例确定原型、建立 this、传参执行初始化，并按构造返回规则决定结果。大写命名只是约定，不会让不可构造函数自动变成构造器。

```js
function User(name) { this.name = name; }
User.prototype.read = function () { return this.name; };
const user = new User('Alice');
console.log(user.read()); // Alice
console.log(Object.getPrototypeOf(user) === User.prototype); // true
```

## 显式 return 会发生什么

对这里的普通基类函数构造：返回对象或函数时，以该对象为 new 的结果；返回原始值（包括 null）或没写 return，通常继续返回已创建实例。

```js
function ReturnsObject() { this.value = 1; return { value: 2 }; }
function ReturnsNumber() { this.value = 1; return 100; }
function ReturnsNull() { this.value = 1; return null; }
console.log(new ReturnsObject().value); // 2
console.log(new ReturnsNumber().value); // 1
console.log(new ReturnsNull().value); // 1
```

显式返回另一个对象后，该结果不一定通过 instanceof 原构造函数。不能把“使用过 new”当成返回对象一定有指定原型的证明。

## 可构造能力与普通调用

箭头、async 函数、普通方法简写通常不能用 new；类构造器需要 new，不能用普通函数调用代替。绑定函数即使没有自有 prototype，也可能在目标可构造时被 new 调用。

因此视频“每个函数都有 prototype”的概括不成立。构造能力是内部语义，不能仅靠属性存在与否判断。

忘记 new 调普通构造函数时，this 会按普通调用规则决定，可能是 undefined 或全局对象，不会自动提供新实例。需要强制构造时可用 class，或在常规函数里检查 new.target。

以上步骤是入门模型；派生 class 的 this 初始化与返回值规则有额外要求，不能不加区分套用。原型背景见 [[前端面试 P029：对象原型、prototype 与属性查找]]。
