---
title: "前端面试 P029：对象原型、prototype 与属性查找"
kind: "concept"
tags: ["前端面试", "JavaScript", "原型", "原型链"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "b8d0fe44-3eb4-5d32-baf8-d0a6327da08d"
---

来源：[方格前端讲面试 · 29. 什么是原型？什么是原型链？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=29)，16 分 33 秒。

回看定位：[00:45 原型定义](https://www.bilibili.com/video/BV14Ttw6WEkt?p=29&t=45) · [04:29 prototype 属性](https://www.bilibili.com/video/BV14Ttw6WEkt?p=29&t=269) · [07:52 方法与实例](https://www.bilibili.com/video/BV14Ttw6WEkt?p=29&t=472) · [12:00 查找规则](https://www.bilibili.com/video/BV14Ttw6WEkt?p=29&t=720)

## 分清两种不同的关联

对象有内部 `[[Prototype]]` 关联，可用 Object.getPrototypeOf 读取；许多可构造函数另有一个普通属性 `.prototype`，默认构造实例时常用它作为新对象的原型。

二者不能混为一谈。`Object.getPrototypeOf(Constructor)` 查构造函数对象自身继承谁，`Constructor.prototype` 查它用来关联常规实例的原型对象。

```js
function User(name) { this.name = name; }
User.prototype.read = function () { return this.name; };
const first = new User('Alice');
const second = new User('Bob');
console.log(Object.getPrototypeOf(first) === User.prototype); // true
console.log(first.read === second.read); // true
console.log(first.read()); // Alice
console.log(Object.getPrototypeOf(User) === Function.prototype); // true
```

## 属性查找与遮蔽

普通读取先看自身，没有才沿原型逐级查找，直到 null。子对象的同名自有属性可遮蔽原型属性；调用继承来的普通方法时，this 通常仍是调用接收者，而不是定义方法的原型。

修改共享的原型对象能影响继续沿该链查找的实例；把 Constructor.prototype 整体换成另一个对象，不会自动把已经存在的实例重新接到新对象上。

## 对视频概括的边界补充

- 对象不必通过构造函数创建，Object.create(proto) 可直接指定原型，因此不能对所有对象都说“原型必然等于创建它的构造函数的 prototype”。
- 并非每个函数都有自有 prototype，例如箭头函数；有 prototype 也不保证可用 new，例如生成器函数。
- Object.create(null) 的链不会经过 Object.prototype。
- __proto__ 是历史遗留的访问器，不是内部槽本身；它有标准兼容背景，但仍不推荐作为常规操作接口。

Object.getPrototypeOf 在 ES5 已有，Object.setPrototypeOf 是 ES2015；版本细节不影响这里推荐使用标准 API 的结论。尽量在创建时决定原型，避免在热点路径频繁改链。

关联：[[前端面试 P030：构造函数与 new 的返回规则]]、[[前端面试 P033：Object.create、描述符与共享属性]]。
