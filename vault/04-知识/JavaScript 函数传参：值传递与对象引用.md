---
title: "JavaScript 函数传参：值传递与对象引用"
kind: "concept"
tags: ["JavaScript", "函数", "值传递", "对象引用"]
created: "2026-10-26T05:51:34Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "06cd5433-0c68-5627-87f7-8bc739b0e779"
---

## 核心结论

JavaScript 的函数参数按值传递。形参是独立的局部变量，不是调用方变量的别名。

对于对象，参数得到的是与实参相同的对象值。可以把这个过程理解为“复制一份指向对象的引用”：没有复制整个对象，所以函数内外能够访问同一个对象；两个变量又彼此独立，所以给形参重新赋值不会改变外部变量的指向。

## 原题与结果

```js
function modify(obj) {
  obj.name = 'Modified';
  obj = { name: 'New Object' };
}

let myObj = { name: 'Initial' };
modify(myObj);

console.log(myObj.name); // 'Modified'
```

## 按执行顺序理解

1. `let myObj = { name: 'Initial' }` 创建对象 A，让 `myObj` 指向它。
2. `modify(myObj)` 初始化局部参数 `obj`。`obj` 和 `myObj` 是两个变量，但此时指向同一个对象 A。
3. `obj.name = 'Modified'` 修改对象 A 的属性。通过 `myObj` 读取同一个对象，也会看到这个变化。
4. `obj = { name: 'New Object' }` 创建对象 B，并让局部参数 `obj` 指向对象 B。这一步没有给 `myObj` 赋值。
5. 函数结束后，`myObj` 仍然指向对象 A，因此输出 `'Modified'`。

刚进入函数时：

```text
myObj ──→ 对象 A { name: 'Initial' }
obj   ──→ 对象 A（同一个对象）
```

执行 `obj.name = 'Modified'` 后：

```text
myObj ──→ 对象 A { name: 'Modified' }
obj   ──→ 对象 A（同一个对象）
```

执行 `obj = { name: 'New Object' }` 后、函数返回前：

```text
myObj ──→ 对象 A { name: 'Modified' }
obj   ──→ 对象 B { name: 'New Object' }
```

| 操作 | 修改的内容 | 外部变量的表现 |
| --- | --- | --- |
| `obj.name = 'Modified'` | 共享对象的属性 | `myObj.name` 读到 `'Modified'` |
| `obj = { name: 'New Object' }` | 局部参数 `obj` 的绑定 | `myObj` 继续指向原对象 |

## 为什么传入对象引用，仍然叫按值传递

容易产生的疑问是：“传进去的明明是引用，为什么还叫值传递？”

这里有两个不同层面：“传入的内容是什么”，以及“参数通过什么机制接收它”。

| 说法 | 含义 |
| --- | --- |
| 传入对象的引用值 | 两个独立变量指向同一个对象 |
| 按引用传递参数 | 形参成为调用方变量的别名，操作的是同一个变量 |

“按值传递”强调复制值，不要求这个值必须是数字或字符串。即使把对象引用作为值传入，也不意味着形参成为外部变量的别名。

可以把原题调用时的关键效果理解成下面这次赋值；这只是解释模型，不是函数调用的完整展开：

```js
let obj = myObj;
```

这会初始化一个独立的变量 `obj`。给它重新赋值，不等于给 `myObj` 重新赋值。

如果是严格意义上的按引用传递，`obj` 就相当于 `myObj` 的另一个名字。那么 `obj = { name: 'New Object' }` 也会改变 `myObj`，原题应输出 `'New Object'`。JavaScript 实际输出 `'Modified'`，说明形参与外部变量并不是这种别名关系。

修改共享对象的属性，并不能证明参数按引用传递。对象的这种共享行为，也常被称为按共享传递（call by sharing）。

## 引用、地址与对象的区别

- 对象是被访问的那份数据，例如 `{ name: 'Initial' }`。
- 引用可以类比为找到对象的地址；它不是对象本身。
- 两个变量可以指向同一个对象，不需要再增加一层“对象值的引用才是真正的对象”。

“复制地址”只是帮助理解的模型。JavaScript 不直接向普通代码暴露对象的实际内存地址；描述参数传递时，也不需要依赖引擎如何安排内存。

## 基本类型同样按值传递

```js
function change(num) {
  num = 100;
}

let count = 10;
change(count);

console.log(count); // 10
```

参数 `num` 最初得到数值 `10`。给局部变量 `num` 重新赋值，不会改变外部变量 `count`。原题中给 `obj` 重新赋值也是一样的道理，只是对象参数还具有共享对象这一特点。

## 如何让外部变量接收新对象

函数返回新对象，再由调用方明确赋值：

```js
function modify(obj) {
  obj.name = 'Modified';
  obj = { name: 'New Object' };
  return obj;
}

let myObj = { name: 'Initial' };
myObj = modify(myObj);

console.log(myObj.name); // 'New Object'
```

这里改变 `myObj` 指向的是调用方的 `myObj = modify(myObj)`，并不是函数内部给参数赋值的操作。

## 来源

2026-10-26 的 JavaScript 学习讨论，根据原题和关于值传递、对象引用的追问整理。

本题也出现在这套课程的 [P3 · 08:51 函数参数传递机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=3&t=531)。本集关于对象赋值、const 与相等比较的完整笔记见 [[前端面试 P003：原始值、对象共享与赋值行为]]。
