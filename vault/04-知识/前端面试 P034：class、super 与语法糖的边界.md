---
title: "前端面试 P034：class、super 与语法糖的边界"
kind: "concept"
tags: ["前端面试", "JavaScript", "class", "super", "继承"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "b5226a07-b2df-51d3-9ade-f4875dcb1683"
---

来源：[方格前端讲面试 · 34.class 是语法糖吗？背后发生了什么？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=34)，14 分 02 秒。

回看定位：[04:35 基础类](https://www.bilibili.com/video/BV14Ttw6WEkt?p=34&t=275) · [07:36 super 的原要求](https://www.bilibili.com/video/BV14Ttw6WEkt?p=34&t=456) · [10:50 语法糖](https://www.bilibili.com/video/BV14Ttw6WEkt?p=34&t=650)

## class 仍然使用原型，但不只是文本缩写

class 定义构造行为、原型方法、静态成员及字段等。普通方法一般位于实例原型上，static 成员在类本身，实例字段与箭头函数字段则在实例上，不能一概说类体所有方法都共享于 prototype。

```js
class User {
  constructor(name) { this.name = name; }
  read() { return this.name; }
  static type() { return 'User'; }
}
const user = new User('Alice');
console.log(user.read()); // Alice
console.log(User.type()); // User
console.log(Object.keys(User.prototype)); // []
```

类方法默认不可枚举，类体严格模式，类构造器不能直接普通调用，声明绑定初始化前有 TDZ。私有字段等还有专门语义，不能把所有 class 特性完全还原为简单的函数加几次赋值。

## super 不必是第一条语句

视频说派生构造函数必须“第一行调用 super”，过于绝对。关键是访问 this 之前要完成相应的 super 初始化；不访问 this 的参数处理可以先执行。

```js
class Base { constructor(name) { this.name = name; } }
class Child extends Base {
  constructor(name) {
    const normalized = name.trim();
    super(normalized);
  }
}
console.log(new Child(' Alice ').name); // Alice
```

super.method() 查找父侧方法，但普通调用中的接收者仍是当前实例。super() 走父构造语义，不等价于把 Parent.call(this) 写得短一点，尤其涉及 class、内建子类与 new.target 时。

## 类型与设计

extends 同时影响实例方法链和类对象的静态继承关联。类方法并不会自动绑定 this，取出后调用仍要考虑接收者。

“语法糖”可以帮助理解它继承既有原型机制，却不应成为忽略这些语义差异的理由。需要复用行为时，还应根据耦合关系比较组合与继承。

关联：[[前端面试 P035：继承方式对照与寄生组合式继承]]、[[前端面试 P029：对象原型、prototype 与属性查找]]。
