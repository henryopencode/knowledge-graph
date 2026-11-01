---
title: "前端面试 P002：数据类型与 typeof 的陷阱"
kind: "concept"
tags: ["JavaScript", "前端面试", "数据类型", "typeof", "类型判断"]
created: "2026-10-26T07:42:21Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "92e0de74-ab42-5f1e-8773-b673b713c7c6"
---

来源：[方格前端讲面试 · P2：数据类型和 typeof 的陷阱](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2)，10 分 49 秒。

回看定位：[01:23 数据类型](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2&t=83) · [04:07 typeof null](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2&t=247) · [06:42 判断数组](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2&t=402) · [07:32 通用判断的说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2&t=452) · [09:08 未声明变量](https://www.bilibili.com/video/BV14Ttw6WEkt?p=2&t=548)

## JavaScript 的类型与 typeof 的返回值

JavaScript 有七种原始类型：Undefined、Null、Boolean、Number、BigInt、String、Symbol，以及 Object 类型。数组、函数、日期等属于对象。

`typeof` 是一元操作符，结果是字符串。它的返回值并不是语言类型的一对一名称表：`null` 和函数就是需要特别理解的例子。

| 值或表达式 | typeof 的结果 | 注意点 |
| --- | --- | --- |
| undefined | "undefined" | 原始值 |
| null | "object" | 历史遗留行为；null 本身属于 Null 类型 |
| true | "boolean" | 原始值 |
| 123、NaN | "number" | NaN 仍是 Number 类型的值 |
| 123n | "bigint" | 不等同于 number |
| "hello" | "string" | 返回值也是字符串 |
| Symbol() | "symbol" | 原始值 |
| {}、[]、new Date() | "object" | 无法靠 typeof 细分这些对象 |
| function () {} | "function" | 函数仍属于对象 |

```js
console.log(typeof null); // object
console.log(typeof []); // object
console.log(typeof NaN); // number
console.log(typeof function () {}); // function
console.log(typeof typeof 123); // string
```

## typeof null 为什么是 object

视频用早期实现中的类型标签与空指针表示解释这个历史问题。需要记住的是当前语言行为：`typeof null === 'object'`，但 `null` 不是一个可以访问属性的普通对象。

判断空值时可以直接用 `value === null`。如果要判断一个非空对象，常见条件是 `value !== null && typeof value === 'object'`；若业务也接收函数，需要另外考虑函数。

历史实现只解释这个结果的来源，不意味着当前所有引擎仍使用同一套底层内存表示。

## 判断数组优先用 Array.isArray

```js
console.log(Array.isArray([])); // true
console.log(Array.isArray({})); // false
console.log(Array.isArray({ length: 0 })); // false
```

`typeof []` 只返回 `'object'`。`value instanceof Array` 依赖具体构造函数及原型链，在 iframe 等不同 realm 之间可能判断失败；`Array.isArray` 更适合确认一个值是不是数组。

## Object.prototype.toString.call 的用途与边界

```js
const tag = value => Object.prototype.toString.call(value);

console.log(tag(null)); // [object Null]
console.log(tag([])); // [object Array]
console.log(tag({})); // [object Object]
console.log(tag('hello')); // [object String]
```

它能为常见值提供比 `typeof` 更细的标签。不过视频在约 07:32–08:47 将它描述为可以精确覆盖所有类型的通用方案，这需要纠正：结果可能受到 `Symbol.toStringTag` 的影响。

```js
const fakeArray = { [Symbol.toStringTag]: 'Array' };

console.log(Object.prototype.toString.call(fakeArray)); // [object Array]
console.log(Array.isArray(fakeArray)); // false
```

因此，不能因为标签是 `[object Array]` 就在任意输入上保证它是真数组。针对数组应使用 `Array.isArray`；针对其他能力，应使用相应 API 或业务校验规则。这个反例说明，视频中“精确覆盖所有类型”的绝对说法不成立。

## typeof 对未声明变量的特殊处理与 TDZ

```js
console.log(typeof aNameThatWasNeverDeclared); // undefined

try {
  {
    console.log(typeof value); // ReferenceError
    let value = 1;
  }
} catch (error) {
  console.log(error.name); // ReferenceError
}
```

第一种情况没有对应绑定，`typeof` 特殊地返回字符串 `'undefined'`。第二种情况存在词法绑定，但还在暂时性死区，依然会抛错。不能记成“typeof 在任何情况下都不会报错”。

暂时性死区的原因见 [[前端面试 P001：var、let、const 与暂时性死区]]。

## 选用哪种判断方式

| 需要判断什么 | 优先考虑 |
| --- | --- |
| 基本类型或可调用函数 | typeof，同时处理 null 等特例 |
| 是否恰好为 null | value === null |
| 是否是数组 | Array.isArray(value) |
| 是否是 Number 类型的 NaN | Number.isNaN(value) |
| 常见内建标签 | Object.prototype.toString.call，注意可定制标签 |
| 业务对象能否使用 | 验证业务需要的字段、类型和能力 |

## 参考与关联

- [MDN：typeof](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/typeof)
- [MDN：Array.isArray](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Array/isArray)
- [MDN：Object.prototype.toString](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Object/toString)
- [[前端面试 P001：var、let、const 与暂时性死区]]
