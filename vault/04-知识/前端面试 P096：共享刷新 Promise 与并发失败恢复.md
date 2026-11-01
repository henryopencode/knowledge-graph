---
title: "前端面试 P096：共享刷新 Promise 与并发失败恢复"
kind: "concept"
tags: ["前端面试", "前端工程", "Token", "并发", "Promise"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "f9ecf41a-ee73-4dc9-9273-d75fea288108"
---

来源：[方格前端讲面试 · 前端如何实现token的无限刷新](https://www.bilibili.com/video/BV14Ttw6WEkt?p=96)，21 分 15 秒。

## 用同一 Promise 让等待者共享结果

布尔 refreshing 只能说明忙不忙，等待者还需要知道何时成功、失败。保存进行中的刷新 Promise，可以自然让所有请求 await 同一个结果。

```js
function createRefreshGate(refresh, save) {
  let pending = null;
  return function refreshOnce() {
    if (!pending) {
      pending = Promise.resolve()
        .then(refresh)
        .then(tokens => { save(tokens); return tokens; })
        .finally(() => { pending = null; });
    }
    return pending;
  };
}
```

refresh 是实际续期函数，save 保存新凭据。抛错或拒绝会传给所有等待者，finally 释放门控，让后续允许的请求有明确状态；不是自己再创建一组无人 reject 的队列 Promise。

## 接入请求层的策略

过期请求先确认它是否已经用过旧令牌；如果当前凭据已被其他请求刷新，可能只需用新值重试，不应因晚到的旧 401 再无意义刷新一次。

刷新后重试需使用新认证头并保留原参数、取消信号等条件，最多重放一次。等待期间用户取消、退出登录或超时，应让相应请求退出，不能刷新成功后把所有旧意图复活。

若仍采用显式请求队列，成功和失败都要排空并分别 resolve/reject；在循环里处理队列新增也要避免覆盖或丢失项。

## 这个片段没有实现什么

它只展示单运行上下文的共享 Promise；跨标签页锁、刷新令牌轮换、服务端撤销、重试幂等与退出版本检查需要额外协议。不要把它单独作为完整认证系统。

关联：[[前端面试 P096：令牌无感刷新、重试与安全边界]]、[[前端面试 P024：Promise 并发组合与空输入边界]]。
