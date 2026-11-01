---
title: "前端面试 P032：instanceof、原型链与自定义判断"
kind: "concept"
tags: ["前端面试", "JavaScript", "instanceof", "Symbol.hasInstance"]
created: "2026-10-26T09:07:55Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "2795695e-6037-578b-bb39-2f3c994c8604"
---

来源：[方格前端讲面试 · 32. instanceof 背后的原理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=32)，21 分 00 秒。

回看定位：[04:25 默认原理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=32&t=265) · [17:16 跨 realm](https://www.bilibili.com/video/BV14Ttw6WEkt?p=32&t=1036) · [18:57 自定义](https://www.bilibili.com/video/BV14Ttw6WEkt?p=32&t=1137)

## 默认关系检查

常规情况下，value instanceof Constructor 检查 Constructor.prototype 是否位于 value 的原型链上。查找从 value 的原型开始，不是从 value 自己开始。

```js
function User() {}
const user = Object.create(User.prototype);
console.log(user instanceof User); // true
console.log('hello' instanceof String); // false
console.log(new String('hello') instanceof String); // true
console.log(Object.create(null) instanceof Object); // false
```

第一个结果说明：instanceof 不证明对象历史上真的由该构造函数创建。原型链可被构造或修改，constructor 属性也可被改写。

## 原始值与自定义 Symbol.hasInstance

对常规构造函数，原始值不会进入普通对象的原型链查找，通常直接为 false；不是先对 null 读原型导致异常再返回 false。

但右侧可定制 Symbol.hasInstance，连原始值也能得到 true：

```js
const Even = {
  [Symbol.hasInstance](value) {
    return typeof value === 'number' && value % 2 === 0;
  },
};
console.log(2 instanceof Even); // true
console.log(3 instanceof Even); // false
```

因此视频“无论什么构造器，对原始值都为 false”的说法需要限定默认行为。右侧也不能毫无条件用普通对象代替构造器；若没有有效定制逻辑或构造器条件不满足，可能抛 TypeError。

## 跨 realm 与类型判断

iframe 或 vm 的内建构造器身份可能不同，外部数组不一定 instanceof 当前 Array。判断真数组优先 Array.isArray；Object.prototype.toString 可被 Symbol.toStringTag 定制，不是无条件可靠替代。

方法调用里 this 的品牌检查、私有字段等还可能比原型关联更严格。伪造原型后通过 instanceof，不代表就能使用所有内建或私有方法。

关联：[[前端面试 P002：数据类型与 typeof 的陷阱]]、[[前端面试 P029：对象原型、prototype 与属性查找]]。依据：[MDN：instanceof](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/instanceof)。
