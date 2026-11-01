---
title: "前端面试 P077：key、组件身份与状态重置"
kind: "concept"
tags: ["前端面试", "React", "key", "组件身份"]
created: "2026-10-26T10:10:10Z"
updated: "2026-10-26T10:10:10Z"
category: "前端"
document_id: "254d7909-edd3-473e-b73e-86deb09b450d"
---

来源：[方格前端讲面试 · 7.React中的key属性有什么作用？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=77)，15 分 38 秒。

回看定位：[00:26 身份](https://www.bilibili.com/video/BV14Ttw6WEkt?p=77&t=26) · [00:46 索引](https://www.bilibili.com/video/BV14Ttw6WEkt?p=77&t=46) · [11:53 重置](https://www.bilibili.com/video/BV14Ttw6WEkt?p=77&t=713)

## key 参与同级匹配

同一父级下，类型与 key 等因素共同决定复用身份，状态随组件身份保留。key 不是全站唯一 ID，也不会自动作为普通 prop 传入子组件；业务 ID 要另外传。

```jsx
function List({ rows }) {
  return rows.map(row => <Editor key={row.id} row={row} />);
}
```

动态插入、删除、排序时，稳定业务 ID 通常比索引更能保留正确输入和局部状态。随机数、每次渲染的时间戳会制造新身份，引起反复卸载挂载。

## 不是无条件的性能开关

视频把某次插入描述成固定的“创建一次、移动两次，所以一定更快”，不能作为一般算法计数。实际操作取决于位置、类型、协调实现与内容；无 key 也可能就地更新而不是全部销毁。

key 首先保证期望的身份语义，性能应具体分析。静态无状态列表里索引可能足够，不需要把三个苛刻条件当作所有应用的唯一允许证明。

## 有意重置

```jsx
<UserForm key={currentUser.id} user={currentUser} />
```

切换身份时 key 变化会重置这部分组件状态，并运行相应卸载清理与新挂载。它不会自动清空外部 store、取消未配合清理的请求或重置整个应用。

组件类型改变、在不稳定位置定义组件等也会影响保留，排查状态丢失不要只盯 key。

依据：[React：保留和重置状态](https://react.dev/learn/preserving-and-resetting-state)。关联：[[前端面试 P075：props 只读、状态归属与派生数据]]。
