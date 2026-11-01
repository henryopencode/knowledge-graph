---
title: "前端面试 P023：async、await、异步恢复与并发"
kind: "concept"
tags: ["前端面试", "JavaScript", "async", "await", "并发"]
created: "2026-10-26T09:01:31Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "cb2c6532-7755-54a9-b088-a3eaa42bc423"
---

来源：[方格前端讲面试 · 23. async、await 的底层运行机制](https://www.bilibili.com/video/BV14Ttw6WEkt?p=23)，22 分 15 秒。

回看定位：[12:42 不阻塞线程](https://www.bilibili.com/video/BV14Ttw6WEkt?p=23&t=762) · [21:37 并发](https://www.bilibili.com/video/BV14Ttw6WEkt?p=23&t=1297)

## async 与 await 分别做什么

async 函数调用返回 Promise；正常 return 的值被用于解析该 Promise，未捕获异常导致拒绝。函数会同步执行到首次 await 等暂停位置，并不是整个函数体一开始就进异步队列。

await 等待值或 thenable 的结果，并在后续异步恢复当前函数。它不阻塞整个 JavaScript 线程，但 await 后面的同步重计算仍会占用线程。

## 纠正：await 普通值也会暂停

视频说 await 非 Promise 值“不会暂停、立即继续”，不正确。即使 await 10，后续也会让出当前同步执行过程，之后按异步恢复规则继续。

```js
async function run() {
  console.log('A');
  await 10;
  console.log('C');
}
run();
console.log('B');
// 输出：A、B、C
```

## 顺序等待不等于必须顺序启动

await loadA(); await loadB() 会等 A 后再调用 B，适合依赖关系。独立任务可先启动，或使用 Promise.all 同时接入错误处理。

```js
async function loadBoth(loadA, loadB) {
  const [a, b] = await Promise.all([loadA(), loadB()]);
  return { a, b };
}
```

如果先单独创建两个 Promise，却长时间只 await 第一个，第二个先拒绝时可能暂时没有处理；统一组合可以避免这种悬空处理窗口。Promise.all 不自带并发数量限制，也不会取消其他任务。

## 运行环境与底层模型

现代 ES 模块可以使用顶层 await，不能只记“await 永远只能写在 async 函数里”。但普通脚本、普通非 async 函数并不会因此获得这项语法。

视频把 async/await 说成必须依赖 Generator 的实现，这是教学或转译模型，不是现代引擎必须真的创建 Generator 对象的规范要求；关键是 Promise 与暂停恢复语义。

fetch 成功拿到 Response 不代表 HTTP 状态成功。404、500 通常仍 fulfilled，要检查 response.ok；response.json 又是一次可能失败的异步读取与解析。

依据：[MDN：await](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/await)。关联：[[前端面试 P024：Promise 并发组合与空输入边界]]。
