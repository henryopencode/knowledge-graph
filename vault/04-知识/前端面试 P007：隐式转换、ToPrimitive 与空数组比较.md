---
title: "前端面试 P007：隐式转换、ToPrimitive 与空数组比较"
kind: "concept"
tags: ["前端面试", "JavaScript", "类型转换", "ToPrimitive"]
created: "2026-10-26T07:55:54Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "7352f106-8f61-59e8-a0f5-f24b843a3cfb"
---

来源：[方格前端讲面试 · 7.隐式类型转换有哪些坑？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=7)，18 分 40 秒。

回看定位：[03:41 ToPrimitive](https://www.bilibili.com/video/BV14Ttw6WEkt?p=7&t=221) · [04:59 转换顺序的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=7&t=299) · [09:03 空数组比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=7&t=543) · [16:31 falsy 列表](https://www.bilibili.com/video/BV14Ttw6WEkt?p=7&t=991)

## 隐式转换发生在哪里

算术、关系比较、宽松相等、条件判断等场景有不同转换规则。不要用一句“都先转换成数字”解释所有操作，尤其要区分对象转原始值与布尔判断。

## 对象转原始值的顺序：需要纠正视频

视频约 04:59 说“没有明确期望时默认先 toString”，不适用于普通对象。准确过程是：

1. 先查找可用的 Symbol.toPrimitive 方法，它可能来自原型链；以 'number'、'string' 或 'default' 作为 hint 调用。
2. 没有该方法时，普通对象在 hint 为 number 或 default 时先 valueOf，返回的仍是对象才尝试 toString。
3. hint 为 string 时先 toString，再 valueOf。两者都不能得到原始值时抛 TypeError。

Date 等对象有专门的转换行为。数组在默认情况下并不是直接跳过 valueOf；而是它的 valueOf 返回数组本身，随后 toString 才给出字符串。

```js
const obj = { valueOf: () => 7, toString: () => 'text' };
console.log(obj + 1); // 8
console.log(String(obj)); // text
```

## [] == ![] 为什么为 true

```text
[] == ![]
[] == false    // 空数组是 truthy，取反得到 false
[] == 0        // 宽松比较把布尔值转成数字
'' == 0        // 数组转原始值，得到空字符串
0 == 0         // 字符串转数字
true
```

空数组在条件判断中为真，并不意味着 `[] == true` 为真；布尔上下文和 == 使用不同算法。

## 加号与显式转换

```js
console.log(1 + '2'); // '12'
console.log(true + true); // 2
console.log(1 + null); // 1
console.log(1 + undefined); // NaN
console.log(Number('')); // 0
console.log(Number('12px')); // NaN
console.log(parseInt('12px', 10)); // 12
```

加号先把对象转成原始值，存在字符串时做拼接；否则按数值规则计算。BigInt 和 Number 不能直接混合加法。Number 与 parseInt 的用途也不同：前者要求整体转换，后者可以解析开头的数字片段。

## falsy 列表不能漏掉 BigInt

视频用“六个 falsy 值”概括，需补充：false、undefined、null、0、-0、0n、NaN、空字符串。0 与 -0 常被合并讲解，但 0n 是独立需要注意的 BigInt 值。

普通对象和数组均为 truthy，布尔转换不会先调用它们的 valueOf 或 toString。浏览器还有 document.all 这个历史特例；日常业务不要利用它设计判断逻辑。

## 实际使用

对输入先明确验证和转换，使用 === 表达严格比较；不要靠背怪题替代理解各操作的规则。相等比较单独见 [[前端面试 P008：相等比较、NaN 与 Object.is]]。

依据：[MDN：Symbol.toPrimitive](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Symbol/toPrimitive)、[MDN：Falsy](https://developer.mozilla.org/en-US/docs/Glossary/Falsy)。
