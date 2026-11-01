---
title: "前端面试 P004：浅拷贝、深拷贝与 structuredClone"
kind: "concept"
tags: ["JavaScript", "前端面试", "浅拷贝", "深拷贝", "structuredClone"]
created: "2026-10-26T07:55:54Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "f1bfee81-15a3-5a35-bbe4-43fce4fa9e69"
---

来源：[方格前端讲面试 · P4：深浅拷贝的本质（含常见实现）](https://www.bilibili.com/video/BV14Ttw6WEkt?p=4)，16 分 15 秒。本集的手写部分另见 [[前端面试 P004：手写深拷贝、循环引用与实现边界]]。

## 先区分赋值与拷贝

```js
const original = { a: 1, nested: { count: 2 } };
const same = original;

console.log(same === original); // true
same.nested.count = 20;
console.log(original.nested.count); // 20
```

这里没有创建另一个对象。对象赋值只让两个变量访问同一个对象，见 [[前端面试 P003：原始值、对象共享与赋值行为]]。

## 浅拷贝只复制一层属性值

```js
const original = { a: 1, nested: { count: 2 } };
const copy = { ...original };

copy.a = 10;
copy.nested.count = 20;

console.log(copy === original); // false
console.log(copy.nested === original.nested); // true
console.log(original.a); // 1
console.log(original.nested.count); // 20
```

外层对象不同，但 `nested` 属性仍指向同一个子对象。原始值属性重新赋值不会影响另一份对象；通过共享的子对象修改属性则会相互可见。

常见浅拷贝包括对象展开 `{ ...obj }`、`Object.assign({}, obj)`，以及数组的 `[...array]`、`array.slice()` 等。它们的细节并不完全一致，不能只因为都“浅”就认为行为相同。例如对象展开和 `Object.assign` 对目标对象 setter 的处理有区别。

## JSON 序列化不等于通用深拷贝

`JSON.parse(JSON.stringify(value))` 适合满足 JSON 数据约束的场景，不适合任意 JavaScript 值。

```js
const original = {
  date: new Date('2026-01-01T00:00:00Z'),
  missing: undefined,
  fn: () => 1,
  symbol: Symbol('id'),
  number: NaN,
};

const copy = JSON.parse(JSON.stringify(original));
console.log(typeof copy.date); // string
console.log(Object.hasOwn(copy, 'missing')); // false
console.log(Object.hasOwn(copy, 'fn')); // false
console.log(Object.hasOwn(copy, 'symbol')); // false
console.log(copy.number); // null
```

还要注意：

- 对象属性里的 `undefined`、函数、Symbol 值通常被忽略；数组中的相应位置通常变成 `null`。
- 循环引用会抛错。
- 没有定制序列化处理时，BigInt 会抛错。
- Map、Set、原型及属性描述符等不能通过这套过程完整保留。
- `toJSON` 等自定义行为也会影响结果。

## structuredClone 适合支持的可克隆数据

```js
const original = {
  nested: { count: 2 },
  date: new Date('2026-01-01T00:00:00Z'),
};
original.self = original;

const copy = structuredClone(original);
copy.nested.count = 20;

console.log(original.nested.count); // 2
console.log(copy.self === copy); // true
console.log(copy.date instanceof Date); // true
console.log(copy.date.getTime() === original.date.getTime()); // true
```

结构化克隆能处理循环引用，以及 Date、Map、Set、ArrayBuffer 等支持的类型，因此通常比 JSON 往返更适合复制这些数据。

## 需要纠正：Error 对象可以结构化克隆

视频约 09:11 的口述和课件把 Error 对象列为不能克隆的类型，这不正确。标准支持 Error 类型，name、message 等错误信息可以被保留；自定义扩展字段是否保留不能据此一概保证。

```js
const copy = structuredClone(new TypeError('bad input'));
console.log(copy instanceof TypeError); // true
console.log(copy.message); // bad input
```

视频最后还建议复杂场景用 Lodash.cloneDeep“拷贝函数”。这里也不能理解为复制出一个行为独立的新函数：克隆库通常不会重建函数闭包；例如对象属性中的函数可能继续共享同一个函数对象。选库仍需要确认它支持的类型和语义。

## 对“深拷贝首选”的边界补充

视频推荐现代 JavaScript 中优先考虑 `structuredClone`，这个方向对支持的数据很实用，但不能理解为“任何对象都能完整复制”。

- 函数、DOM 节点等不能直接结构化克隆，可能抛 `DataCloneError`。
- 自定义类的原型、私有字段、访问器和属性描述符等，不会按普通对象的全部行为原样重建。
- 复制对象图时，会保留副本内部原本的共享关系，而不是为每个属性重复制造一份不同对象。
- 传输 ArrayBuffer 会使源缓冲区分离；SharedArrayBuffer 的底层内存仍可共享。这些是具体类型的语义，不能一概背成“所有内存完全独立”。

因此，先确定数据类型、是否需要保留类行为、是否有循环引用，再选择工具。更新某个字段时，也不一定需要深拷贝整个对象。

## 面试回答思路

浅拷贝创建新的外层对象，但嵌套对象仍可能共享；深拷贝需要根据支持的数据类型递归或按相应算法复制对象图，并处理循环引用。对象展开、`Object.assign` 是常见浅拷贝，JSON 往返有数据丢失和循环引用等限制，`structuredClone` 能处理多种内建类型但也有明确边界。

## 参考与关联

- [MDN：structuredClone](https://developer.mozilla.org/en-US/docs/Web/API/Window/structuredClone)
- [MDN：结构化克隆算法](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API/Structured_clone_algorithm)
- [MDN：JSON.stringify](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/JSON/stringify)
- [[前端面试 P004：手写深拷贝、循环引用与实现边界]]
