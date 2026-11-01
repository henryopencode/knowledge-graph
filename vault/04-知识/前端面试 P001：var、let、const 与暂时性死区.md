---
title: "前端面试 P001：var、let、const 与暂时性死区"
kind: "concept"
tags: ["JavaScript", "前端面试", "变量声明", "作用域", "TDZ"]
created: "2026-10-26T07:42:21Z"
updated: "2026-10-26T09:57:12Z"
category: "前端"
document_id: "fd4d54e0-9ea3-5e80-b0c9-b8a5e380f08b"
---

来源：[方格前端讲面试 · P1：var,let,const 有什么区别？](https://www.bilibili.com/video/BV14Ttw6WEkt?p=1)，18 分 59 秒。

回看定位：[01:59 var 的特性](https://www.bilibili.com/video/BV14Ttw6WEkt?p=1&t=119) · [06:22 提升说法与 TDZ](https://www.bilibili.com/video/BV14Ttw6WEkt?p=1&t=382) · [12:52 TDZ](https://www.bilibili.com/video/BV14Ttw6WEkt?p=1&t=772) · [13:39 全局对象](https://www.bilibili.com/video/BV14Ttw6WEkt?p=1&t=819)

## 先理解它们分别约束什么

`var`、`let`、`const` 都用来声明变量。比较时需要分别看作用域、初始化之前能否访问、重复声明、能否重新赋值，而不是只记一句“新语法替代旧语法”。

| 比较项 | var | let | const |
| --- | --- | --- | --- |
| 常见作用域 | 函数作用域；顶层行为取决于脚本或模块环境 | 块级作用域 | 块级作用域 |
| 声明语句执行前访问 | 通常得到 undefined | 暂时性死区内抛 ReferenceError | 暂时性死区内抛 ReferenceError |
| 同一作用域重复声明 | 可以重复 var 声明，但不能因此与同名 let、const 混用 | 不允许 | 不允许 |
| 重新赋值 | 允许 | 允许 | 不允许 |
| 普通声明时是否必须给初始值 | 不必 | 不必 | 必须 |

最后一行讨论 `const value = ...` 这样的普通声明；`for...of`、`for...in` 的循环头有相应的语法规则。

## var 的函数作用域与提前读取

```js
function example() {
  console.log(value); // undefined
  var value = 10;

  if (true) {
    var value = 20;
  }

  console.log(value); // 20
}

example();
```

这里的两个 `var value` 指向同一个函数作用域里的变量。`if` 的花括号不会替 `var` 建立独立的变量作用域，所以块内赋值后，块外读到的也是 `20`。

所谓“变量提升”，不是引擎真的把源代码搬到上面。理解这个例子时，只需要知道：函数体开始执行前，该 `var` 绑定已经建立并初始化为 `undefined`，但 `value = 10` 仍在原来位置执行。

## let 的块级作用域

```js
let value = 10;

if (true) {
  let value = 20;
  console.log(value); // 20
}

console.log(value); // 10
```

两个 `value` 属于不同作用域。内层变量遮蔽外层变量，修改内层变量不会给外层变量赋值。

同一作用域内重复 `let` 声明会产生语法错误；嵌套作用域里再次声明同名变量是另一回事。

## 暂时性死区 TDZ

从进入相应作用域，到 `let` 或 `const` 绑定完成初始化之前，访问该绑定会抛出 `ReferenceError`。

```js
let value = 'outer';

try {
  {
    console.log(value); // ReferenceError
    let value = 'inner';
  }
} catch (error) {
  console.log(error.name); // ReferenceError
}
```

内层作用域里的 `value` 从一开始就会遮蔽外层变量，因此第一行读取不会退回去拿 `'outer'`。

视频在约 06:22 和 16:28 将 let、const 概括为“没有变量提升”，这个术语表述不够严谨。“提升”不是严格的规范术语；更准确、也更容易预测结果的说法是：词法绑定已建立，但尚未初始化，因此不能访问。

TDZ 的结束点是完成初始化，不只是代码走到声明语句的开头。例如 `let value = value` 在右侧读取同一个未初始化绑定，仍会抛出 `ReferenceError`。

## const 限制绑定，不会自动冻结对象

```js
const user = { name: 'Alice' };
user.name = 'Bob';
console.log(user.name); // Bob

try {
  user = { name: 'Carol' };
} catch (error) {
  console.log(error.name); // TypeError
}
```

修改 `user.name` 是修改对象属性；`user = ...` 是给变量重新赋值。`const` 禁止后者，不自动禁止前者。这与 [[JavaScript 函数传参：值传递与对象引用]] 中“改属性”和“改变量绑定”的区别一致。

## 对视频中全局对象说法的适用范围补充

视频用 `window.globalVar` 演示顶层 `var` 与 `let` 的区别。这个例子需要明确运行环境：

- 在浏览器普通脚本中，顶层 `var` 通常会建立对应的全局对象属性；顶层 `let`、`const` 不会因此成为 `window` 属性。
- 在 ES 模块中，顶层 `var` 也是模块作用域，不会因此成为 `window` 属性。
- Node.js CommonJS 文件也有自己的模块作用域，不能照搬浏览器普通脚本的结论。

这属于补充适用条件，不影响课件中普通浏览器脚本示例的结果。不要把“在顶层声明 var”一概等同于“在所有环境中都给 window 添加属性”。

## 面试时可以怎样回答

`var` 主要是函数作用域，绑定会预先初始化为 `undefined`，可以重复 `var` 声明和重新赋值。`let`、`const` 是块级作用域，初始化前存在暂时性死区，同一作用域不能重复声明。`let` 可以重新赋值，`const` 不可以，但 `const` 指向的对象仍可能被修改。实际开发中，绑定不需要重新赋值时优先使用 `const`，需要重新赋值时使用 `let`。

## 参考与关联

- [MDN：var](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/var)
- [MDN：let](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/let)
- [MDN：const](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/const)
- [[JavaScript 函数传参：值传递与对象引用]]
