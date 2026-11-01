---
title: "前端面试 P088：Immer 草稿与结构共享"
kind: "concept"
tags: ["前端面试", "React", "Redux", "Immer"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "0fb4aa98-de11-4d83-8d23-87977ab59850"
---

来源：[方格前端讲面试 · 18.为什么说 Immer.js 在现代 Redux 生态中扮演着如此重要的角色？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=88)，11 分 07 秒。

回看定位：[02:50 痛点](https://www.bilibili.com/video/BV14Ttw6WEkt?p=88&t=170) · [04:02 produce](https://www.bilibili.com/video/BV14Ttw6WEkt?p=88&t=242) · [06:31 结构共享](https://www.bilibili.com/video/BV14Ttw6WEkt?p=88&t=391) · [08:14 RTK](https://www.bilibili.com/video/BV14Ttw6WEkt?p=88&t=494)

## 只修改草稿，不修改输入状态

produce(base, recipe) 让 recipe 接收 draft，以近似可变赋值的形式表达修改，再生成保持不可变快照语义的结果。被修改路径产生新对象，未改部分尽量共享原引用，不是全量深拷贝。

```js
import { produce } from 'immer';
const previous = { user: { name: 'Alice' }, settings: { theme: 'light' } };
const next = produce(previous, draft => { draft.user.name = 'Bob'; });
console.log(previous.user.name); // Alice
console.log(next.user.name); // Bob
console.log(previous.user === next.user); // false
console.log(previous.settings === next.settings); // true
console.log(produce(previous, () => {}) === previous); // true
```

## 不能因此声称任意对象都安全克隆

draft 有作用域，不应泄漏到外面异步继续修改。recipe 不要一边改 draft 一边返回另一个新值；Map/Set、自定义类等有具体支持条件，不是所有对象都可无差别代理。

结构共享以外的冻结、代理和复制也有成本。课程称它兼顾体验与效率方向正确，但不保证每个场景都更快，更不能把它当作任意数据通用深克隆器。

## Redux Toolkit 的集成

createSlice 等相应 reducer 场景由 Immer 包装，所以写 state.field = value 可以是合法草稿修改；离开该包装直接修改已有 Redux state 仍错误。不要把相同语法跨边界照搬。

“每次 action 必然产生新快照对象”也不准确：没有修改时可以返回同一引用，正是示例最后一行展示的行为。

依据：[Immer：produce](https://immerjs.github.io/immer/produce/)。关联：[[前端面试 P004：浅拷贝、深拷贝与 structuredClone]]。
