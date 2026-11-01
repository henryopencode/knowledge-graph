---
title: "前端面试 P067：Vue 模板编译与运行时边界"
kind: "concept"
tags: ["前端面试", "Vue", "编译", "AST", "render"]
created: "2026-10-26T10:04:02Z"
updated: "2026-10-26T10:04:02Z"
category: "前端"
document_id: "e25012c2-b225-4799-bc01-cd4dfccf7e56"
---

来源：[方格前端讲面试 · 25.Vue模板编译流程](https://www.bilibili.com/video/BV14Ttw6WEkt?p=67)，7 分 02 秒。

回看定位：[00:34 三个阶段](https://www.bilibili.com/video/BV14Ttw6WEkt?p=67&t=34) · [03:09 静态优化](https://www.bilibili.com/video/BV14Ttw6WEkt?p=67&t=189) · [04:20 代码生成](https://www.bilibili.com/video/BV14Ttw6WEkt?p=67&t=260)

## 模板不是每次直接交给浏览器重新解析

模板经解析形成 AST，经过相应转换与优化，再生成渲染代码。渲染函数运行后描述虚拟节点，运行时再负责挂载与更新。常见单文件组件会在构建阶段预编译，生产运行时不必再次从模板源码编译。

```vue
<template>
  <section><h1>固定标题</h1><p>{{ message }}</p></section>
</template>
```

这个片段中固定部分与动态文本提供了不同优化机会。实际输出与 Vue/编译器版本、选项有关，不能仅看示意伪代码就声称它等同当前源码。

## Vue 3 的优化不限于静态节点

静态提升或缓存、patch flag、block tree、动态属性分析等都能减少运行时工作。parse → transform → codegen 是更贴近 Vue 3 编译主线的描述；课程用独立“静态标记优化阶段”帮助入门，但忽略了其他转换。

## 需要纠正版本混合

视频提到通过 with(this) 生成渲染函数，这是 Vue 2 等旧模式常见形式；Vue 3 现代构建输出通常采用明确上下文与辅助函数等形式，不应把旧式 with 模板作为所有 Vue 版本的固定实现。

模板表达式也不是随意执行服务端代码的通道。不要把不可信用户内容当模板编译；HTML 字符串显示、v-html 与可信模板编译有不同的安全边界。

手写极简 parser 可能不处理完整 HTML、指令、插值和错误恢复。它适合说明 AST，不是可替换官方编译器的实现。

依据：[Vue：渲染机制](https://vuejs.org/guide/extras/rendering-mechanism.html)。关联：[[前端面试 P068：Vue 2 与 Vue 3 的列表 diff 区别]]。
