---
title: "前端面试 P045：ref、reactive、computed 与 watch"
kind: "concept"
tags: ["前端面试", "Vue", "响应式", "ref", "computed", "watch"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "d8ca2e82-a0f5-4b69-a343-7aa5823ac5f7"
---

来源：[方格前端讲面试 · 3.Composition-API基础语法讲解](https://www.bilibili.com/video/BV14Ttw6WEkt?p=45)，11 分 30 秒。

回看定位：[00:58 ref](https://www.bilibili.com/video/BV14Ttw6WEkt?p=45&t=58) · [03:59 computed](https://www.bilibili.com/video/BV14Ttw6WEkt?p=45&t=239) · [06:44 watch](https://www.bilibili.com/video/BV14Ttw6WEkt?p=45&t=404)

## 四个 API 分工

| API | 用途 | 注意 |
| --- | --- | --- |
| ref | 保存响应式值，包括原始值或对象 | 脚本通常经 .value 访问；对象默认也可深层响应 |
| reactive | 为对象建立响应式代理 | 直接访问属性，不能把原始值当可代理对象 |
| computed | 从响应式依赖派生值并缓存 | getter 尽量纯；也可显式提供 setter |
| watch | 观察数据源变化并执行副作用 | 区分 ref、getter、代理对象等数据源与调度配置 |

```js
import { ref, reactive, computed, watch, nextTick } from 'vue';
const count = ref(1);
const state = reactive({ name: 'Alice' });
const double = computed(() => count.value * 2);
watch(() => state.name, (next, previous) => console.log(previous, next));
count.value++;
console.log(double.value); // 4
state.name = 'Bob';
await nextTick(); // watcher 打印 Alice Bob
```

监听属性时用 `() => state.name`，而不是把当时读出的字符串传给 watch。深度监听原地修改对象时，新旧值可能是同一代理，并非自动保存修改前快照。

## computed 的缓存与可写形式

computed 依赖所追踪的响应式读取，通常惰性求值，依赖变化使缓存需要重新校验；不能把“每次赋值立即重算 getter”当成规则。

getter-only computed 默认只读。写入通常产生 Vue 警告并被忽略，不能笼统保证一定抛出运行时异常。需要可写计算属性时提供 get/set，并清晰定义它怎样更新源数据。

## 对视频 ref 分类的补充

ref 不只支持数字、字符串，也能保存对象，且支持整体替换 .value。reactive 适合对象，但不是对象状态唯一合法选择。API 选择要看替换方式、解构与状态归属，不要把“常用”背成“只能”。

watch 的 immediate 可让回调开始时执行，deep 会影响遍历成本；需要 DOM 更新后观察时使用合适的 flush 或 nextTick，而不是猜测时序。

依据：[Vue：响应式基础](https://vuejs.org/guide/essentials/reactivity-fundamentals.html)、[Vue：侦听器](https://vuejs.org/guide/essentials/watchers.html)。
