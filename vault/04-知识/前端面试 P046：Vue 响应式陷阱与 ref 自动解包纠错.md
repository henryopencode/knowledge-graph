---
title: "前端面试 P046：Vue 响应式陷阱与 ref 自动解包纠错"
kind: "concept"
tags: ["前端面试", "Vue", "响应式", "ref", "解构"]
created: "2026-10-26T09:48:01Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "c3af600f-d7aa-4066-979e-2255585d470e"
---

来源：[方格前端讲面试 · 4.Vue的响应式陷阱](https://www.bilibili.com/video/BV14Ttw6WEkt?p=46)，9 分 17 秒。

回看定位：[00:40 解构](https://www.bilibili.com/video/BV14Ttw6WEkt?p=46&t=40) · [04:05 替换 ref 的原建议](https://www.bilibili.com/video/BV14Ttw6WEkt?p=46&t=245) · [07:31 嵌套 ref 的原建议](https://www.bilibili.com/video/BV14Ttw6WEkt?p=46&t=451)

## 原始值属性解构通常失去持续关联

从 reactive 对象读取 name 字符串，解构出来的是当时的值，不再自动读取原代理。需要保持连接时可以用 toRef/toRefs，或一直通过原对象访问。

```js
import { reactive, ref, toRefs } from 'vue';
const user = reactive({ name: 'Alice' });
const { name } = toRefs(user);
name.value = 'Bob';
console.log(user.name); // Bob
```

若解构得到的是一个代理子对象，继续修改它的属性仍可能响应；不能把“任何解构都彻底不响应”当成统一规则。

## 必须纠正：普通 reactive 对象属性会解包 ref

课件约 07:35 推荐对 `reactive({ count: ref(0) })` 写 state.count.value++，这是错误用法。对普通深层 reactive 对象的属性，ref 会自动解包：

```js
import { reactive, ref } from 'vue';
const count = ref(0);
const state = reactive({ count });
console.log(state.count); // 0
state.count++;
console.log(count.value); // 1
```

数组元素或 Map 值等位置不同，通常不会这样自动解包：

```js
import { reactive, ref } from 'vue';
const list = reactive([ref(1)]);
console.log(list[0].value); // 1
```

shallowReactive 等 API 还有不同规则，需要按具体容器确认。

## 必须纠正：整体替换 ref.value 是受支持的

```js
import { ref } from 'vue';
const state = ref({ count: 0 });
const old = state.value;
state.value = { count: 1 };
console.log(state.value.count); // 1
console.log(old.count); // 0
```

旧引用继续指向旧对象，是 JavaScript 对象身份的正常语义，不代表 ref 的替换失效。需要跟随替换的数据应持续读取 state.value，而不是一次拿出旧值后期待它自动改指向。

Object.assign 合并和整体替换表达不同意图：合并会保留未覆盖字段，删除语义也不同。不能只为保留旧引用而一概禁止替换。

## Vue 2 与 Vue 3 不要混记

Vue 3 的代理可追踪新增属性与数组索引写入；不必为迁就旧版本默认改写全部逻辑。splice 适合插入删除，不一定比直接赋某个索引更清晰。

Vue 3.5 的 defineProps 解构有编译器支持，但它是特定语法转换，不等于普通 JavaScript reactive 解构都被修复。

依据：[Vue：ref 解包细节](https://vuejs.org/guide/essentials/reactivity-fundamentals.html#additional-ref-unwrapping-details)。关联：[[前端面试 P003：原始值、对象共享与赋值行为]]。
