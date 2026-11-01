---
title: "前端面试 P091：SWR、TanStack Query 与服务端缓存"
kind: "concept"
tags: ["前端面试", "React", "SWR", "TanStackQuery", "缓存"]
created: "2026-10-26T10:15:35Z"
updated: "2026-10-26T10:15:35Z"
category: "前端"
document_id: "a01cac57-96f0-4bfd-9a5a-9bb2560ab402"
---

来源：[方格前端讲面试 · 21.像 SWR 或 React Query 这样的库，它们主要解决了什么问题？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=91)，15 分 02 秒。

回看定位：[05:53 客户端与服务端](https://www.bilibili.com/video/BV14Ttw6WEkt?p=91&t=353) · [07:07 缓存策略](https://www.bilibili.com/video/BV14Ttw6WEkt?p=91&t=427) · [11:01 去重的原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=91&t=661) · [12:51 高级场景](https://www.bilibili.com/video/BV14Ttw6WEkt?p=91&t=771)

## 获取结果只是同步生命周期的一部分

服务端数据有远端所有者，可能过期、被他人修改或因网络失败暂时不可用。SWR、TanStack Query 等帮助管理查询状态、缓存、重新验证、失效、重试、分页和 mutation，而不只是替换 fetch。

## key 决定缓存身份

```js
const queryOptions = {
  queryKey: ['article', articleId, locale],
  queryFn: async ({ signal }) => {
    const response = await fetch(`/api/articles/${articleId}?locale=${locale}`, { signal });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
  },
};
```

这是 TanStack Query 风格的配置片段。影响结果的 ID、语言、筛选或会话范围需要体现在 key/缓存策略中；漏掉输入会把不同请求当作同一份结果。退出登录还要考虑清理或隔离私有缓存。

## 纠正“总是只请求一次、永远保持最新”

同 key 去重和缓存复用有条件、时间窗与库配置，不保证所有挂载永远只发一次请求。staleTime、gcTime、焦点和网络重连策略影响重新获取；SWR 与 TanStack Query 的默认值也不完全相同。

旧缓存加后台验证改善体验，但不是强一致或实时推送保证。乐观更新需定义失败回滚和并发写冲突，重试也不应盲目重复非幂等操作。

## 卸载不等于自动泄漏

视频把卸载后 setState 概括为内存泄漏，这是过度判断。真正问题是无用资源仍被持续保留、过期请求写回等具体行为；现代 React 也已移除部分旧警告。库的取消能力还取决于 fetcher 是否使用信号。

缓存库是客户端的数据同步工具，不会自动消除服务端权限问题，也不是 HTTP Cache-Control 的完整等价替代。

依据：[TanStack Query 重要默认值](https://tanstack.com/query/latest/docs/framework/react/guides/important-defaults)、[SWR：重新验证](https://swr.vercel.app/docs/revalidation)。
