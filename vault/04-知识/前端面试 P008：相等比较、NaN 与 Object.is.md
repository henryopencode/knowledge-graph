---
title: "前端面试 P008：相等比较、NaN 与 Object.is"
kind: "concept"
tags: ["前端面试", "JavaScript", "相等比较", "NaN", "Object.is"]
created: "2026-10-26T07:55:54Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "887f45b3-0f21-5eb2-a386-409353c4a58e"
---

来源：[方格前端讲面试 · 8.== 和 === 的核心考点](https://www.bilibili.com/video/BV14Ttw6WEkt?p=8)，10 分 10 秒。

回看定位：[02:04 严格比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=8&t=124) · [02:11 宽松比较](https://www.bilibili.com/video/BV14Ttw6WEkt?p=8&t=131) · [08:38 NaN](https://www.bilibili.com/video/BV14Ttw6WEkt?p=8&t=518)

## === 与 == 的核心区别

=== 不做类型转换；类型不同直接为 false。同类型也要遵守相应规则：对象比较身份，NaN 不等于自身，+0 与 -0 相等。

== 在不同类型间按规范分支处理，不是简单地先给双方套 Number 或 Boolean。同类型时通常采用与严格相等相同的结果。

```js
console.log(77 === '77'); // false
console.log(77 == '77'); // true
console.log(true == 1); // true
console.log(null == undefined); // true
console.log(null == 0); // false
console.log([] == false); // true
console.log([] === false); // false
```

对象与原始值比较时，可能触发对象转原始值，详见 [[前端面试 P007：隐式转换、ToPrimitive 与空数组比较]]。两个对象比较不会自动深度比较内容。

```js
const a = {};
const b = a;
console.log(a === b); // true
console.log(a === {}); // false
console.log(a == {}); // false
```

## nullish 判断

对普通值，`value == null` 常用于同时匹配 null 和 undefined，相当于 `value === null || value === undefined`。浏览器 document.all 是历史例外，不适合在这里概括为毫无例外的等价。

=== 是业务代码更直观的默认选择，但“永远只能使用 ===”属于编码约定，不能替代对 == 行为的理解。

## NaN、+0、-0 与不同相等算法

| 表达式 | 结果 |
| --- | --- |
| NaN === NaN | false |
| Number.isNaN(NaN) | true |
| Number.isNaN('not a number') | false |
| Object.is(NaN, NaN) | true |
| +0 === -0 | true |
| Object.is(+0, -0) | false |
| [NaN].includes(NaN) | true |
| [NaN].indexOf(NaN) | -1 |

Object.is 采用 SameValue 语义；Array.includes、Set、Map 的相关键值比较使用 SameValueZero，能把 NaN 看作与自身相同，并将 +0、-0 视为相同。

## 视频内容与补充的区别

视频关于 === 不转换类型、对象按身份比较、用 Number.isNaN 判断 NaN 的主要结论正确。这里额外补充 Object.is、SameValueZero 和 document.all 边界，避免把“NaN 永远与自己不相等”误套到所有 API。

依据：[MDN：Equality comparisons and sameness](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Equality_comparisons_and_sameness)。
