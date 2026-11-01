---
title: "前端面试 P068：Vue 2 与 Vue 3 的列表 diff 区别"
kind: "concept"
tags: ["前端面试", "Vue", "diff", "key", "LIS"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "cc31e77a-1084-4e10-ada4-6700501e5328"
---

来源：[方格前端讲面试 · 26.Vue的diff算法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=68)，8 分 08 秒。

回看定位：[02:00 四端算法原说法](https://www.bilibili.com/video/BV14Ttw6WEkt?p=68&t=120) · [04:19 key](https://www.bilibili.com/video/BV14Ttw6WEkt?p=68&t=259) · [05:12 就地复用](https://www.bilibili.com/video/BV14Ttw6WEkt?p=68&t=312)

## 匹配节点与复用身份

虚拟 DOM 更新需要判断可复用节点、变化的属性文本、新增删除和顺序移动。相同 key 不是唯一条件，节点或组件类型也影响是否复用；不同父级的相同 key 也不等于跨层迁移同一实例。

## 必须纠正：Vue 3 不等于 Vue 2 的四端比较

视频把 Vue 2 与 Vue 3 都概括为同一套双端四指针 diff、并统一说成 O(n)。Vue 2 经典算法常讲新旧头尾四端比较；Vue 3 keyed children 常见流程是同步前后缀、处理中间未匹配段，建立 key 映射并在需要移动时利用最长递增子序列减少移动。

LIS 部分通常涉及 O(n log n) 的计算，实际工作还包括递归 patch 与 DOM 操作；不能给整个渲染过程无条件套一个统一 O(n) 结论。

## key 的首要价值是身份正确

稳定 ID 能使有状态输入、组件实例跟随对应数据条目。使用随机 key 会频繁重新挂载；动态重排时用索引可能让状态跟着位置而非业务对象。

但无 key 的就地更新并不等于必然“全部销毁重建”或永远更慢；纯展示内容可能只改文本。课程把 key 只解释为性能加速过于单一，状态与生命周期的正确性通常更重要。

虚拟 DOM 也不是所有场景都比直接 DOM 操作快，它提供声明式更新、比较与调度的组织能力，性能还受编译优化和应用结构影响。

## 复习时记流程而不是错误口号

先比较是否可复用，再处理新旧子节点；key 在同级范围帮助匹配，实际移动策略按具体版本理解。不要把“复杂树编辑算法 O(n³)”的背景介绍等同于框架每次真的先做了那种算法再优化。

依据：[Vue：渲染机制](https://vuejs.org/guide/extras/rendering-mechanism.html)。关联：[[前端面试 P052：Vue 条件、列表与表单指令]]。
