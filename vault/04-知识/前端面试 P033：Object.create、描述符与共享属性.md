---
title: "前端面试 P033：Object.create、描述符与共享属性"
kind: "concept"
tags: ["前端面试", "JavaScript", "Object.create", "属性描述符"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "1ff329d4-3a8c-5efc-9ebf-39e5f7ab212e"
---

来源：[方格前端讲面试 · 33. Object.create 是怎么实现继承的？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=33)，13 分 43 秒。

回看定位：[04:16 API](https://www.bilibili.com/video/BV14Ttw6WEkt?p=33&t=256) · [08:21 无原型](https://www.bilibili.com/video/BV14Ttw6WEkt?p=33&t=501) · [11:54 描述符](https://www.bilibili.com/video/BV14Ttw6WEkt?p=33&t=714)

## 创建关联，不调用父构造函数

Object.create(proto, descriptors) 建立一个新对象，并直接使用指定对象或 null 作为原型；它不自动复制父对象，也不运行父构造器。

```js
const animal = {
  type: 'animal',
  describe() { return this.type; },
};
const dog = Object.create(animal);
dog.type = 'dog';
console.log(dog.describe()); // dog
console.log(animal.type); // animal
console.log(Object.getPrototypeOf(dog) === animal); // true
```

dog 自有 type 遮蔽原型属性，继承的方法依然以 dog 为接收者运行。

## 无原型对象：null 与 __proto__ 不同

```js
const dictionary = Object.create(null);
dictionary.name = 'Alice';
console.log(Object.getPrototypeOf(dictionary)); // null
console.log(dictionary.__proto__); // undefined
console.log(Object.hasOwn(dictionary, 'name')); // true
```

课件通过 obj.__proto__ 得到 undefined 并据此说明原型，容易混淆。真正的内部原型是 null；undefined 是因为该对象没有继承 __proto__ 访问器。也不要直接调用 dictionary.hasOwnProperty，它通常不存在。

## 第二参数是描述符，默认并非可枚举

```js
const obj = Object.create(null, { value: { value: 1 } });
console.log(obj.value); // 1
console.log(Object.keys(obj)); // []
console.log(Object.getOwnPropertyDescriptor(obj, 'value').writable); // false
```

视频说可用第二参数定义自身“可枚举”属性，需要纠正为按描述符决定：没有写 enumerable、writable、configurable 时默认都是 false。

## 继承也可能共享可变对象

若原型含 items 数组，child.items.push(x) 修改的是查找到的共享数组，会影响其他通过同一原型访问它的对象。只有 child.items = newArray 才通常建立子对象自己的绑定属性。

所以 Object.create 不是隔离父对象影响的克隆机制；只描述属性查找与委派关系。共享状态背景见 [[前端面试 P003：原始值、对象共享与赋值行为]]。
