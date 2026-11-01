---
title: "前端面试 P017：手写 bind、重复绑定与构造调用"
kind: "concept"
tags: ["前端面试", "JavaScript", "bind", "手写代码", "new"]
created: "2026-10-26T08:49:37Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "fe79636e-1965-56d1-a261-8eb2b4d78878"
---

来源：[方格前端讲面试 · 17. 手写 bind 实现](https://www.bilibili.com/video/BV14Ttw6WEkt?p=17)，23 分 25 秒。

回看定位：[04:53 实现目标](https://www.bilibili.com/video/BV14Ttw6WEkt?p=17&t=293) · [10:38 构造判断](https://www.bilibili.com/video/BV14Ttw6WEkt?p=17&t=638) · [20:09 重复绑定的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=17&t=1209)

## 应保留的核心行为

返回新函数；普通调用使用指定 thisArg；预设参数排在后续参数前；构造调用时忽略 thisArg，仍传递预设参数。箭头目标不能因为绑定后就获得构造能力。

## 需要纠正：再次 bind 不是完全无效

视频课件说第二次 bind 无法改变第一次绑定的 this 和参数。准确的是：不能改掉第一层保存的 this，但第二次仍会返回新函数，并且可以继续添加预设参数。

```js
function describe(...args) { return [this.name, ...args]; }
const first = describe.bind({ name: 'A' }, 1);
const second = first.bind({ name: 'B' }, 2);
console.log(second(3)); // ['A', 1, 2, 3]
console.log(first === second); // false
```

## 一个明确有限的教学包装器

```js
function bindBasic(fn, receiver, ...prefix) {
  if (typeof fn !== 'function') throw new TypeError('Expected function');
  function bound(...suffix) {
    const args = [...prefix, ...suffix];
    if (new.target) {
      return Reflect.construct(fn, args, new.target === bound ? fn : new.target);
    }
    return Reflect.apply(fn, receiver, args);
  }
  return bound;
}

function User(name) { this.name = name; }
const receiver = { name: 'old' };
const BoundUser = bindBasic(User, receiver, 'Alice');
const user = new BoundUser();
console.log(user.name); // Alice
console.log(receiver.name); // old
console.log(user instanceof User); // true
```

使用 new.target 区分构造调用，比 `this instanceof bound` 更直接。后者可能被 `bound.call(Object.create(bound.prototype))` 这样的普通调用误导。

## 为什么仍不能称为原生 bind 的完整替代

上述普通函数包装器有自己的 prototype，与原生绑定函数不同；`instanceof BoundUser` 也不自动获得原生绑定函数的委派检查语义。目标是否可构造、name、length、原型、new.target、Proxy、类构造函数等都需要进一步处理。

视频的简化实现直接 `new Original(...args)`，能让结果通过 instanceof Original，却不能据此保证 instanceof Bound 同样为 true；课件中的这个预期需要额外实现才能成立。

严格目标函数的原始 thisArg 不会自动装箱；非严格目标才有相应替换规则。业务中使用原生 bind，手写时清楚声明范围。

关联：[[前端面试 P014：call、apply、bind 与模拟实现边界]]。依据：[MDN：bind](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Function/bind)。
