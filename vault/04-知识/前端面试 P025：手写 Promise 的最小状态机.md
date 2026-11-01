---
title: "前端面试 P025：手写 Promise 的最小状态机"
kind: "concept"
tags: ["前端面试", "JavaScript", "Promise", "手写代码"]
created: "2026-10-26T09:01:32Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "748ba177-7538-5b8f-8419-d05712bff102"
---

来源：[方格前端讲面试 · 25. 手写一个简化版 Promise](https://www.bilibili.com/video/BV14Ttw6WEkt?p=25)，24 分 36 秒。

回看定位：[04:38 状态机要素](https://www.bilibili.com/video/BV14Ttw6WEkt?p=25&t=278) · [09:01 异步调度](https://www.bilibili.com/video/BV14Ttw6WEkt?p=25&t=541) · [13:31 简化范围](https://www.bilibili.com/video/BV14Ttw6WEkt?p=25&t=811)

## 先声明这个版本的范围

下面是按课程思路整理的教学状态机：支持同步 executor、单向状态转换、异步反应、普通值链式传递与异常传播。为控制复杂度，**不实现 thenable 吸收、自解析循环拒绝、finally 和静态组合方法**，不能替代原生 Promise 或声称完整符合规范。

```js
class BasicPromise {
  #state = 'pending';
  #value;
  #reactions = [];

  constructor(executor) {
    if (typeof executor !== 'function') throw new TypeError('Expected executor');
    const settle = (state, value) => {
      if (this.#state !== 'pending') return;
      this.#state = state;
      this.#value = value;
      this.#flush();
    };
    try {
      executor(value => settle('fulfilled', value),
        reason => settle('rejected', reason));
    } catch (error) {
      settle('rejected', error);
    }
  }

  #flush() {
    if (this.#state === 'pending') return;
    for (const reaction of this.#reactions.splice(0)) {
      queueMicrotask(() => reaction(this.#state, this.#value));
    }
  }

  then(onFulfilled, onRejected) {
    return new BasicPromise((resolve, reject) => {
      this.#reactions.push((state, value) => {
        const handler = state === 'fulfilled' ? onFulfilled : onRejected;
        if (typeof handler !== 'function') {
          if (state === 'fulfilled') resolve(value);
          else reject(value);
          return;
        }
        try { resolve(handler(value)); }
        catch (error) { reject(error); }
      });
      this.#flush();
    });
  }

  catch(onRejected) { return this.then(undefined, onRejected); }
}

new BasicPromise(resolve => resolve(1))
  .then(value => value + 1)
  .then(value => console.log(value)); // 2
console.log('sync'); // 先输出 sync，再输出 2
```

## 为什么 then 必须返回新实例

每个处理函数的结果属于下一段链：普通返回值决定成功，throw 决定失败，缺少处理函数时值或原因穿透。当前 Promise 一旦 settled 不再改变，下一段仍有自己的结果。

## 对课程调度模型的修正

课程用 setTimeout(0) 模拟异步，这能说明“不要同步调用回调”，但它不是微任务，不能保留与原生 Promise 相同的任务顺序。这里使用 queueMicrotask 展示调度差异。

Promises/A+ 对 then 回调的异步要求，与 ECMAScript 原生 Promise job 的完整调度、构造、解析和静态 API 不是同一整套规范。通过部分示例不能证明完整兼容。

完整实现前还应处理：返回自身、返回其他 Promise、then 属性 getter 抛错、恶意 thenable 多次回调等。解析过程见下一篇 [[前端面试 P025：Promise 解析过程与实现验收]]。
