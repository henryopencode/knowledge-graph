---
title: "前端面试 P022：Promise 状态、链式调用与错误传播"
kind: "concept"
tags: ["前端面试", "JavaScript", "Promise", "错误处理"]
created: "2026-10-26T09:01:31Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "b27a0935-209e-5f47-a90e-bcf960ad7a9e"
---

来源：[方格前端讲面试 · 22. Promise 基本语法 + 错误捕获](https://www.bilibili.com/video/BV14Ttw6WEkt?p=22)，23 分 28 秒。

回看定位：[03:02 状态](https://www.bilibili.com/video/BV14Ttw6WEkt?p=22&t=182) · [04:50 executor](https://www.bilibili.com/video/BV14Ttw6WEkt?p=22&t=290) · [20:08 错误捕获差异](https://www.bilibili.com/video/BV14Ttw6WEkt?p=22&t=1208)

## 状态与 resolve 的区别

Promise 有 pending、fulfilled、rejected 三种状态；fulfilled 或 rejected 后称 settled，最终状态不会再次变化。resolved 不应当简单作为 fulfilled 的同义词：resolve 另一个待定 Promise，会锁定跟随它的结果，但自己可能仍处于 pending。

executor 在构造时立即同步运行。resolve/reject 不会让 executor 像 return 一样自动停止，之后的同步语句仍会继续；首次决定后的后续尝试不能重新改结果。

## then 为什么能串联

then 返回新 Promise。处理函数返回普通值，新链采用该值；返回 Promise 或 thenable，新链跟随它；抛错则拒绝新链。没有相应处理函数时，值或错误继续传递。

```js
Promise.resolve(1)
  .then(value => value + 1)
  .then(() => { throw new Error('step failed'); })
  .catch(error => `recovered: ${error.message}`)
  .then(console.log); // recovered: step failed
```

catch 返回普通值意味着恢复为成功，不是“记录后仍会自动失败”。希望调用方感知失败，需要重新 throw 或返回被拒绝的 Promise。

## 纠正：同一个 then 的失败处理不捕获成功处理抛错

视频比较两种错误处理写法时，口述把 then(success, fail) 的 fail 说成也能捕获同一个 success 的异常，这是错误的。这个异常会拒绝 then 返回的新 Promise，应由下游处理。

```js
Promise.resolve()
  .then(
    () => { throw new Error('from success'); },
    () => console.log('not called'),
  )
  .catch(error => console.log(error.message)); // from success
```

then(success).catch(fail) 位于更下游，因此能处理 success 抛错。catch 自己或后面的 finally 再抛错，仍需要继续处理新链。

## finally 的透明性与边界

视频把 finally 的返回值概括成“被忽略、不影响后续结果”，需限定正常返回。finally 通常不接收成功值或错误原因，适合清理。正常返回通常保持原结果；若它抛错或返回拒绝的 Promise，会以新失败替换原结果。它不是无条件“保证后面成功”的方法。

Promise 自动捕获 executor 同步抛错，不会自动捕获 executor 内另一个定时器回调将来直接 throw 的错误。应在那个异步回调中捕获并调用 reject，且把相关 Promise 返回或等待进同一条链。

依据：[MDN：使用 Promise](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Using_promises)。关联：[[前端面试 P027：async 错误处理与 return await]]。
