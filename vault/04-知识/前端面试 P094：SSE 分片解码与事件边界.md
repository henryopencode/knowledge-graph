---
title: "前端面试 P094：SSE 分片解码与事件边界"
kind: "concept"
tags: ["前端面试", "前端工程", "SSE", "ReadableStream", "编码"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "177b0772-2561-473f-8191-183f9197b52e"
---

来源：[方格前端讲面试 · ChatGPT 的对话功能实现，为什么选择 SSE协议而非Websocket](https://www.bilibili.com/video/BV14Ttw6WEkt?p=94)，10 分 59 秒。

## 解析器不能把每次 read 当完整消息

UTF-8 字符、CRLF 和 JSON 都可能被拆到多个网络 chunk。SSE 以文本行与空行分隔事件，多个 data 字段要拼接换行，冒号开头是注释。

下面是对课程流式场景的补充实现，只提取 data，支持 LF、CR、CRLF；不实现 id、event、retry、重连和完整业务协议。

```js
async function* readEventData(chunks) {
  const decoder = new TextDecoder();
  let line = '', skipLF = false, data = [];
  function acceptLine(text) {
    if (text === '') {
      if (data.length === 0) return null;
      const event = { data: data.join('\n') };
      data = [];
      return event;
    }
    const colon = text.indexOf(':');
    const field = colon < 0 ? text : text.slice(0, colon);
    let value = colon < 0 ? '' : text.slice(colon + 1);
    if (value.startsWith(' ')) value = value.slice(1);
    if (field === 'data') data.push(value);
    return null;
  }
  function* feed(text) {
    for (const character of text) {
      if (skipLF) {
        skipLF = false;
        if (character === '\n') continue;
      }
      if (character === '\r' || character === '\n') {
        const event = acceptLine(line);
        line = '';
        skipLF = character === '\r';
        if (event) yield event.data;
      } else line += character;
    }
  }
  for await (const chunk of chunks) yield* feed(decoder.decode(chunk, { stream: true }));
  yield* feed(decoder.decode());
}
```

连接结束时未以事件边界完成的片段不直接冒充一条完整事件。业务层再根据服务端约定解析 JSON、处理错误或结束标记；不要假设所有 SSE 服务都采用同一个 `[DONE]` 约定。

## 使用边界

处理响应状态与 Content-Type，再读取 response.body。需要最大事件大小、超时、取消、reader 释放和版本一致性；生产中优先使用经过验证的相应协议解析器。

文本增量更新还应避免每个字节触发一次昂贵列表渲染，注意 Markdown 在半个代码块或标签时的显示策略。

关联：[[前端面试 P094：流式文本的 SSE 与 WebSocket 取舍]]。
