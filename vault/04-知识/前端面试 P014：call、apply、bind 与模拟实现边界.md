---
title: "前端面试 P014：call、apply、bind 与模拟实现边界"
kind: "concept"
tags: ["前端面试", "JavaScript", "call", "apply", "bind"]
created: "2026-10-26T08:49:37Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "3605a0b5-5d5e-5ed1-bda2-04f9cf847e0c"
---

来源：[方格前端讲面试 · 14. bind、call、apply的底层原理](https://www.bilibili.com/video/BV14Ttw6WEkt?p=14)，19 分 27 秒。

回看定位：[02:46 call](https://www.bilibili.com/video/BV14Ttw6WEkt?p=14&t=166) · [06:22 apply](https://www.bilibili.com/video/BV14Ttw6WEkt?p=14&t=382) · [14:08 构造调用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=14&t=848) · [16:07 模拟步骤](https://www.bilibili.com/video/BV14Ttw6WEkt?p=14&t=967)

## API 对照

call(thisArg, ...args) 立即调用，apply(thisArg, args) 立即调用并接收类数组参数，bind(thisArg, ...prefix) 返回新函数，延后调用并预置前缀参数。

```js
function greet(prefix, end) {
  return `${prefix}, ${this.name}${end}`;
}
const user = { name: 'Alice' };
console.log(greet.call(user, 'Hi', '!')); // Hi, Alice!
console.log(greet.apply(user, ['Hello', '.'])); // Hello, Alice.
const hello = greet.bind(user, 'Hey');
console.log(hello('?')); // Hey, Alice?
```

apply 的参数是类数组对象，不要求可迭代；展开语法需要 iterable，这两者不能对任意输入互换。参数极多时，两者都可能遇到引擎参数数量限制。

## thisArg 的转换由目标函数规则决定

严格函数保留 null、undefined 和原始 this 值；非严格函数会把 nullish 替换为相应全局对象，并装箱原始值。箭头函数不建立自己的 this，这几个方法不能为它改变词法 this。

bind 是创建包装行为，不是复制函数代码或克隆闭包。new 调用绑定函数时忽略绑定 thisArg，但保留预置参数，这一点本集后半部分也正确提到。

## 临时挂到对象再调用：为什么只是教学模拟

课件利用 obj[symbol](...args) 模拟方法调用，再删除临时属性。这能演示普通对象接收者，但与原生 call/apply 有明显差别：

- 无法忠实表达严格模式下的 null、undefined 或原始值接收者。
- 冻结、不可扩展对象可能拒绝添加属性；Proxy 可能观察或干预这个过程。
- 目标抛错时仍要在 finally 中删除临时属性。
- 修改接收者本身是原生调用不必产生的可观察行为。

业务代码可以直接用 `Reflect.apply(fn, thisArg, args)`；面试需要手写时应先声明模拟范围，不要把 Symbol 临时属性法说成引擎真实实现。

## 保存绑定函数引用

每次 bind 都得到不同的函数对象。给事件监听注册 handler.bind(obj)，移除时再写一次 handler.bind(obj) 不会得到同一个监听器，应该保存第一次返回值。

进一步实现和纠错见 [[前端面试 P017：手写 bind、重复绑定与构造调用]]；this 规则见 [[前端面试 P010：普通函数、箭头函数与 this]]。
