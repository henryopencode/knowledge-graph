---
title: "前端面试 P073：JSX 编译、React 元素与表达式"
kind: "concept"
tags: ["前端面试", "React", "JSX", "编译"]
created: "2026-10-26T10:10:09Z"
updated: "2026-10-26T10:10:09Z"
category: "前端"
document_id: "5088f4bc-215d-4f35-962d-eff6f328a698"
---

来源：[方格前端讲面试 · 3.说说你对JSX的理解](https://www.bilibili.com/video/BV14Ttw6WEkt?p=73)，11 分 56 秒。

回看定位：[02:19 编译](https://www.bilibili.com/video/BV14Ttw6WEkt?p=73&t=139) · [03:57 元素](https://www.bilibili.com/video/BV14Ttw6WEkt?p=73&t=237) · [06:15 表达式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=73&t=375)

## JSX 是语法扩展

JSX 通常需要编译工具转换，浏览器不会直接按普通 JavaScript 解析尖括号组件语法。传统转换可用 React.createElement，自动 JSX runtime 常调用 jsx/jsxs 等辅助函数；并非所有配置都生成相同文本。

```jsx
const items = [{ id: 1, name: 'Alice' }];
const view = <ul>{items.map(item => <li key={item.id}>{item.name}</li>)}</ul>;
```

转换后创建 React 元素描述，不会仅因为给 view 赋值就把节点插进页面，也不会立即执行所有组件函数。需要 React 渲染过程消费它。

## 花括号里是表达式，不是任意语句

可放变量、函数调用、三元、数组映射等；不能直接把 if、for 等语句写在表达式位置。它仍有 JSX 语法规则，不应像视频所说“无限制使用全部 JS”那样理解。

条件 `count && <Badge />` 在 count 为 0 时可能渲染数字 0，若目标是是否有未读消息可写 count > 0。普通对象也不能不加处理直接作为子节点渲染。

## 安全与属性

正常文本插值会转义，不等于整个应用自动免疫 XSS。dangerouslySetInnerHTML、URL、第三方脚本等仍有各自的信任边界。className、事件回调、key 等也有 React 约定。

JSX 不必称为模板来理解，但“语法扩展”是技术分类，不是证明它必然比其他模板系统更强或更快。

依据：[React：JSX 中使用 JavaScript](https://react.dev/learn/javascript-in-jsx-with-curly-braces)。
