---
title: "前端面试 P027：async 错误处理与 return await"
kind: "concept"
tags: ["前端面试", "JavaScript", "async", "错误处理"]
created: "2026-10-26T09:01:32Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "6d911d6f-bb5d-57e5-a377-c54b3671ac30"
---

来源：[方格前端讲面试 · 27. 实战：Async、await 错误处理的 3 种方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=27)，21 分 43 秒。

回看定位：[04:59 三种方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=27&t=299) · [12:45 局部回退](https://www.bilibili.com/video/BV14Ttw6WEkt?p=27&t=765) · [19:14 错误传递](https://www.bilibili.com/video/BV14Ttw6WEkt?p=27&t=1154)

## 三种方式各自负责什么

内部 try/catch 适合有恢复、补偿或转换语义的步骤；调用处的 catch 或外层 await/try 适合统一处理整个操作；单个 Promise.catch 可以提供明确的局部回退。

```js
async function optionalValue() {
  const value = await Promise.reject(new Error('optional failed'))
    .catch(() => 'fallback');
  return value;
}
optionalValue().then(console.log); // fallback
```

catch 返回普通值会恢复为成功；throw 或返回拒绝的 Promise 会继续失败。不要无条件返回 null，除非调用方知道 null 表示可接受的降级结果。

## 不是整个 try 里产生的未来错误都会被抓到

try/catch 能捕获其中同步 throw，以及真正 await 到的拒绝。一个启动后没返回、没 await 的 Promise，或单独 setTimeout 回调将来 throw 的异常，不会因创建时位于 try 里就自动被捕获。

`operation().catch(handler)` 也只处理成功调用 operation 得到的 Promise；如果 operation 在返回之前同步抛错，仍需要外层 try 或明确的 Promise 包装边界。

## return promise 与 return await

```js
async function returnsDirectly() {
  try { return Promise.reject(new Error('failed')); }
  catch { return 'caught'; }
}
async function awaitsInside() {
  try { return await Promise.reject(new Error('failed')); }
  catch { return 'caught'; }
}
returnsDirectly().catch(error => console.log(error.message)); // failed
awaitsInside().then(console.log); // caught
```

直接 return 交给外层跟随 Promise，局部 catch 不会捕获其未来拒绝；return await 则在当前 try 的 await 位置恢复并抛错。因此“return await 永远多余”是错误的经验规则。

## 对视频措辞的精确化

await 拒绝后可以按异常控制流进入 catch，但并不是在 async 函数最初调用时向外同步 throw。async 函数内未捕获异常体现在返回 Promise 的拒绝上，调用者也必须观察它。

finally 中的 return 或 throw 可能覆盖原返回值或原错误，应谨慎使用。错误在合适边界记录和处理，避免每层重复打日志，或每层吞掉错误。

关联：[[前端面试 P022：Promise 状态、链式调用与错误传播]]、[[前端面试 P023：async、await、异步恢复与并发]]。
