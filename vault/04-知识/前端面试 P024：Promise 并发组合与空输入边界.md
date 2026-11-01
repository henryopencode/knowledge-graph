---
title: "前端面试 P024：Promise 并发组合与空输入边界"
kind: "concept"
tags: ["前端面试", "JavaScript", "Promise.all", "Promise.race", "并发"]
created: "2026-10-26T09:01:31Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "ae86fb27-db15-5a74-963a-2fcda14968e8"
---

来源：[方格前端讲面试 · 24. Promise.all 、 allSettled 、race 、any 用法对比](https://www.bilibili.com/video/BV14Ttw6WEkt?p=24)，17 分 46 秒。

回看定位：[02:45 all](https://www.bilibili.com/video/BV14Ttw6WEkt?p=24&t=165) · [07:59 allSettled](https://www.bilibili.com/video/BV14Ttw6WEkt?p=24&t=479) · [09:38 race](https://www.bilibili.com/video/BV14Ttw6WEkt?p=24&t=578) · [12:32 any](https://www.bilibili.com/video/BV14Ttw6WEkt?p=24&t=752)

## 四种方法如何选择

| 方法 | 成功条件 | 失败条件 | 空可迭代输入 |
| --- | --- | --- | --- |
| all | 所有输入成功，值按输入顺序 | 某个输入拒绝 | fulfilled，值为 [] |
| allSettled | 所有输入 settled，返回每项状态 | 普通输入拒绝不会使整体拒绝 | fulfilled，值为 [] |
| race | 跟随最先决定结果的输入 | 最先决定的是拒绝 | 持续 pending |
| any | 最先成功的输入 | 所有输入拒绝，AggregateError | rejected，AggregateError |

all 的结果顺序不是完成顺序；allSettled 的对象为 `{ status: 'fulfilled', value }` 或 `{ status: 'rejected', reason }`；any 的 errors 按输入次序记录原因。

```js
async function demo() {
  console.log(await Promise.all([Promise.resolve(1), 2])); // [1, 2]
  const results = await Promise.allSettled([
    Promise.resolve('ok'), Promise.reject('bad'),
  ]);
  console.log(results.map(result => result.status)); // ['fulfilled', 'rejected']
  console.log(await Promise.any([Promise.reject('bad'), Promise.resolve('ok')])); // ok
  try {
    await Promise.any([]);
  } catch (error) {
    console.log(error instanceof AggregateError); // true
    console.log(error.errors.length); // 0
  }
}
demo();
```

## 成功、失败与取消不同

all 快速失败只决定组合 Promise 的结果，其他请求不会自动停止；race 的失败或成功也不会中止输掉的任务。实现超时时应配合 AbortController，并在结束时清除超时定时器。

这些方法消费可迭代对象，不只是数组。已启动的任务在传入前可能已经执行；Promise.all 不是“调用时才启动所有 Promise”的通用执行器。需要限流时应该保存任务函数，按额度启动。

## 对视频 allSettled 说法的补充

“总会成功”需要限定：每个输入 Promise 的拒绝被转成结果记录，但若获取或遍历输入时抛错，例如 iterable 的迭代器异常，整体仍可能拒绝。空输入则正常成功为 []，并非例外失败。

即使组合 Promise 的状态可立即确定，then/catch 观察结果的回调仍是异步，不能用空数组示例推断 then 同步调用。

关联：[[前端面试 P023：async、await、异步恢复与并发]]。
