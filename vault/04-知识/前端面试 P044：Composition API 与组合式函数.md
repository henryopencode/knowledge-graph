---
title: "前端面试 P044：Composition API 与组合式函数"
kind: "concept"
tags: ["前端面试", "Vue", "CompositionAPI", "组合式函数"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "5cb4fd37-fae9-49c4-babe-65b4b8d62c81"
---

来源：[方格前端讲面试 · 2.Composition-API与响应式系统整体介绍](https://www.bilibili.com/video/BV14Ttw6WEkt?p=44)，10 分 16 秒。

回看定位：[04:22 组织方式](https://www.bilibili.com/video/BV14Ttw6WEkt?p=44&t=262) · [06:18 setup](https://www.bilibili.com/video/BV14Ttw6WEkt?p=44&t=378)

## 逻辑按功能聚合

Options API 按 data、methods、computed、watch 等选项组织；Composition API 可以把同一功能的状态、派生值、监听与清理放在一起，并提取为组合式函数。

```js
import { ref, computed } from 'vue';
function useCounter(initial = 0) {
  const count = ref(initial);
  const double = computed(() => count.value * 2);
  const increment = () => count.value++;
  return { count, double, increment };
}
const first = useCounter();
const second = useCounter();
first.increment();
console.log(first.count.value); // 1
console.log(second.count.value); // 0
```

在函数内创建状态通常使每次调用独立；把 ref 放到模块顶层则可能成为共享状态。这不是 use 前缀自动决定的，SSR 场景尤其要区分每请求与全局状态。

## setup 与 script setup

普通 setup 通过返回对象暴露模板绑定；script setup 是编译语法，顶层绑定可直接供模板使用。它通常按组件实例执行，不是模块只执行一次。

依赖当前实例的生命周期 API 与自动作用域清理，应在有效的 setup 等同步上下文中注册；不能把“都是函数”理解成任何时间、任何地方调用都等价。

## 与 React Hooks 的类比边界

二者都帮助组合逻辑，但 Vue setup 通常不是每次渲染重新执行。ref 不等于 useState，computed 不等于 useMemo，watch 不等于 useEffect；它们的更新、追踪、依赖和生命周期规则不同。

Composition API 不要求替代所有 Options API，已有结构清晰的代码不必仅为语法风格重写。复杂功能拆分的收益应来自内聚和明确接口。

关联：[[前端面试 P045：ref、reactive、computed 与 watch]]、[[前端面试 P012：高阶函数、柯里化与部分应用]]。
