---
title: "前端面试 P010：普通函数、箭头函数与 this"
kind: "concept"
tags: ["前端面试", "JavaScript", "this", "箭头函数", "bind"]
created: "2026-10-26T07:55:54Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "0f55ecae-5bfa-58af-b91a-2e9f2d0807c7"
---

来源：[方格前端讲面试 · 10.JS 中的 this 究竟指向谁？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=10)，20 分 18 秒。

回看定位：[03:30 普通调用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=10&t=210) · [08:47 bind 的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=10&t=527) · [13:11 箭头函数](https://www.bilibili.com/video/BV14Ttw6WEkt?p=10&t=791) · [17:40 定时器的原分析](https://www.bilibili.com/video/BV14Ttw6WEkt?p=10&t=1060)

## this 由什么决定

普通函数的 this 主要取决于调用方式及函数是否严格模式；箭头函数不建立自己的 this，沿创建时的词法环境读取。this 不保证一定是对象，严格函数里可以是 undefined、null 或原始值。

| 调用形式 | 主要规则 |
| --- | --- |
| obj.method() | 普通方法调用通常以 obj 作为 this |
| 独立调用 fn() | 严格函数为 undefined；非严格函数通常替换为全局对象 |
| fn.call(receiver, ...args) | 显式传入 receiver，非严格函数还可能替换 null 或装箱原始值 |
| fn.apply(receiver, args) | 与 call 类似，参数以类数组形式提供 |
| fn.bind(receiver) | 返回绑定函数，普通调用使用保存的 receiver |
| new Constructor() | 通过构造流程建立实例 this，需考虑显式返回对象等规则 |
| 箭头函数 | 使用外层 this，call、apply、bind 不能为它建立自己的 this |

```js
'use strict';
const user = {
  name: 'Alice',
  read() { return this?.name; },
  makeReader() { return () => this.name; },
};
const detached = user.read;
console.log(user.read()); // Alice
console.log(detached()); // undefined
console.log(detached.call({ name: 'Bob' })); // Bob
console.log(user.makeReader().call({ name: 'Carol' })); // Alice
```

对象字面量本身不创建 this 作用域。直接写 `method: () => this.name`，箭头不会因为放在 user 里面就自动获得 user。

## 需要补充：bind 并非任何调用下都永久使用同一个 this

视频对 bind 使用了“无论怎么调用都不会改变”的表述。对普通调用和再次 call/apply 这大体正确，但构造调用是重要边界：当目标可构造时，new 调用绑定函数会忽略绑定的 thisArg，使用新实例。

```js
function User(name) { this.name = name; }
const receiver = { name: 'old' };
const Bound = User.bind(receiver);
const instance = new Bound('new');
console.log(receiver.name); // old
console.log(instance.name); // new
console.log(instance instanceof User); // true
```

箭头函数不能用 new。不要把四条绑定规则的优先级理解为任何语法组合都合法，例如不能直接把 new 与 call 的普通语法拼起来当成通用机制。

## 需要纠正：定时器回调不等于普通独立调用

视频约 17:40 用“定时器独立调用，所以严格时 this 为 undefined”解释回调，这忽略了宿主 API 的规则。

- 浏览器 setTimeout 对普通回调通常以 Window 作为 this，即使回调本身严格模式。
- Node.js 普通定时器回调通常以 Timeout 对象作为 this。
- 箭头函数仍使用外层 this；bind 也能明确固定接收者。

```js
const timer = {
  seconds: 5,
  readLater() {
    return new Promise(resolve => {
      setTimeout(() => resolve(this.seconds), 0);
    });
  },
};
timer.readLater().then(console.log); // 5
```

修复思路是正确的，但原来的原因分析应加上运行环境。DOM 事件监听、Promise 回调等也各有调用语义，不能把所有回调都按“脱离对象所以默认绑定”一概处理。

## 排题顺序

先看是不是箭头；再确认是否构造、已绑定、显式 call/apply，或通过某个接收者调用；最后考虑普通独立调用和严格模式。同时确认是否由定时器或事件系统等宿主 API 调用。

依据：[MDN：bind](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function/bind)、[MDN：setTimeout 的 this 问题](https://developer.mozilla.org/en-US/docs/Web/API/Window/setTimeout#the_this_problem)。
