---
title: "前端面试 P026：Promise 延迟与顺序任务队列"
kind: "concept"
tags: ["前端面试", "JavaScript", "Promise", "任务队列", "取消"]
created: "2026-10-26T09:01:32Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "5158a970-cc1d-5f21-819d-00db14bf69c3"
---

来源：[方格前端讲面试 · 26. 实战：用 Promise 重写 setTimeout 任务队列](https://www.bilibili.com/video/BV14Ttw6WEkt?p=26)，15 分 35 秒。

回看定位：[04:20 延迟包装](https://www.bilibili.com/video/BV14Ttw6WEkt?p=26&t=260) · [09:37 链式 return](https://www.bilibili.com/video/BV14Ttw6WEkt?p=26&t=577) · [12:11 取消](https://www.bilibili.com/video/BV14Ttw6WEkt?p=26&t=731)

## 延迟的最小包装

```js
const delay = ms => new Promise(resolve => setTimeout(resolve, ms));

async function run() {
  console.log('one');
  await delay(0);
  console.log('two');
  await delay(0);
  console.log('three');
}
run();
console.log('outside');
// 输出：one、outside、two、three
```

延迟值只是定时调度条件；Promise 包装不会让定时器变精确，也不会让同步重计算自动并行。

## 顺序执行要保存任务函数

```js
async function runQueue(tasks) {
  const results = [];
  for (const task of tasks) results.push(await task());
  return results;
}
```

传入 `() => request()` 才能控制何时启动。如果在构造数组时已经调用 request 得到 Promise，它可能早已开始，之后逐个 await 不会把启动时机倒回串行。

then 写法也必须 return 本次任务的 Promise，否则下一段链不会等待那份被丢掉的异步工作。

## 取消与错误

原生 Promise 没有统一的 cancel。要取消 delay，可额外接入 AbortSignal，清理计时器并以明确原因拒绝；只 clearTimeout 却不解决那份 Promise，可能使等待者永久 pending。

有实际业务任务时，决定遇错停止还是收集错误继续。后者必须明确返回每项状态，不要默默 catch 掉所有失败，让队列看似全部成功。

视频的基本封装思路正确。它是应用层顺序调度，与浏览器内部的任务队列不是同一个概念，命名相同不意味着能够直接控制宿主队列。

关联：[[前端面试 P024：Promise 并发组合与空输入边界]]、[[前端面试 P020：定时器延迟、轮询重叠与清理]]。
