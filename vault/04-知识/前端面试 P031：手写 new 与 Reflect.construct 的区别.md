---
title: "前端面试 P031：手写 new 与 Reflect.construct 的区别"
kind: "concept"
tags: ["前端面试", "JavaScript", "new", "手写代码", "Reflect"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "e4889877-90ff-537f-9539-0a38c912d777"
---

来源：[方格前端讲面试 · 31. 手写 new 的实现逻辑.](https://www.bilibili.com/video/BV14Ttw6WEkt?p=31)，14 分 44 秒。

回看定位：[06:36 四个步骤](https://www.bilibili.com/video/BV14Ttw6WEkt?p=31&t=396) · [07:23 Object.create](https://www.bilibili.com/video/BV14Ttw6WEkt?p=31&t=443) · [10:24 返回判断](https://www.bilibili.com/video/BV14Ttw6WEkt?p=31&t=624)

## 普通函数的教学模拟

```js
function newBasic(Constructor, ...args) {
  if (typeof Constructor !== 'function') throw new TypeError('Expected function');
  const candidate = Constructor.prototype;
  const prototype = candidate !== null &&
    (typeof candidate === 'object' || typeof candidate === 'function')
    ? candidate : Object.prototype;
  const instance = Object.create(prototype);
  const result = Reflect.apply(Constructor, instance, args);
  return result !== null &&
    (typeof result === 'object' || typeof result === 'function')
    ? result : instance;
}

function User(name) { this.name = name; }
const user = newBasic(User, 'Alice');
console.log(user.name); // Alice
console.log(user instanceof User); // true
```

先建立原型关联，再用新对象调用普通初始化函数。返回函数也是返回对象的情形，判断时不能只写 typeof result === 'object'，也不能把 null 算成对象结果。

## 不能声称完整模拟原生 new

typeof 为 function 不等于可构造。上述版本会对箭头目标作普通调用，而原生 new 应拒绝；class 不能通过 Reflect.apply 初始化；Date 等内建构造器、绑定构造器、new.target、不同 realm 的原型回退也不具备完全相同的行为。

因此 Object.create 加 apply 是普通可调用函数的解释模型，不是通用构造算法。课程没有展开这些差异，笔记在这里补足支持范围。

## 真正需要动态构造时

```js
class User { constructor(name) { this.name = name; } }
const user = Reflect.construct(User, ['Alice']);
console.log(user.name); // Alice
console.log(user instanceof User); // true
```

Reflect.construct 使用目标的构造语义，并可通过第三个 newTarget 控制新对象原型来源等行为。不要把它理解成“只是把 thisArg 换成对象的 apply”。

关联：[[前端面试 P030：构造函数与 new 的返回规则]]、[[前端面试 P017：手写 bind、重复绑定与构造调用]]。
