---
title: "前端面试 P098：用户偏好存储、TTL 与命名空间"
kind: "concept"
tags: ["前端面试", "前端工程", "localStorage", "用户偏好", "TTL"]
created: "2026-10-26T10:33:31Z"
updated: "2026-10-26T10:33:31Z"
category: "前端"
document_id: "66a75614-6a49-4d63-ac70-1d441e156ea1"
---

来源：[方格前端讲面试 · 如何使用 JS 实现一个轻量级本地存储方案，用于保存和读取用户偏好设置](https://www.bilibili.com/video/BV14Ttw6WEkt?p=98)，12 分 29 秒。

回看定位：[00:58 方案选择](https://www.bilibili.com/video/BV14Ttw6WEkt?p=98&t=58) · [02:16 TTL](https://www.bilibili.com/video/BV14Ttw6WEkt?p=98&t=136) · [03:03 命名空间](https://www.bilibili.com/video/BV14Ttw6WEkt?p=98&t=183)

## 选择存储前先限定数据

少量非敏感、可 JSON 序列化的偏好适合 localStorage；会话范围可考虑 sessionStorage，更大或需要事务的数据可考虑 IndexedDB。Web Storage 同步执行，也可能因配额或策略不可用，不保证永不丢失。

## 包装版本、值和到期时间

```js
function createPreferences(storage, namespace = 'prefs') {
  const prefix = `${namespace}:`;
  return {
    set(key, value, ttlMs = null) {
      if (ttlMs !== null && (!Number.isFinite(ttlMs) || ttlMs < 0)) {
        throw new RangeError('Invalid TTL');
      }
      try {
        const record = { version: 1, value,
          expiresAt: ttlMs === null ? null : Date.now() + ttlMs };
        const encoded = JSON.stringify(record);
        if (!Object.hasOwn(JSON.parse(encoded), 'value')) return false;
        storage.setItem(prefix + key, encoded);
        return true;
      } catch { return false; }
    },
    get(key, fallback = null) {
      try {
        const raw = storage.getItem(prefix + key);
        if (raw === null) return fallback;
        const record = JSON.parse(raw);
        if (!record || record.version !== 1 || !Object.hasOwn(record, 'value')) return fallback;
        if (record.expiresAt !== null &&
            (!Number.isFinite(record.expiresAt) || record.expiresAt <= Date.now())) {
          storage.removeItem(prefix + key);
          return fallback;
        }
        return record.value;
      } catch { return fallback; }
    },
    remove(key) { storage.removeItem(prefix + key); },
    clear() {
      const keys = Array.from({ length: storage.length }, (_, index) => storage.key(index));
      for (const key of keys) if (key?.startsWith(prefix)) storage.removeItem(key);
    },
  };
}
```

这是教学版，调用 storage 的 remove/clear 失败会抛错，调用方需按界面错误策略处理。获取 window.localStorage 本身也可能抛错，应在调用入口捕获；对象类型要限制在声明的 JSON 数据范围。

## 需要补足的边界

TTL 通过读时检查近似实现，不是浏览器会主动准点删除。版本迁移、数据校验、容量、当前用户隔离与跨标签页 storage 事件还需设计；同页写入不会靠 storage 事件自动通知同一个页面。

clear 先取键快照，只删本命名空间，避免边遍历边删除漏项，更不能为清偏好直接 localStorage.clear 误删其他功能。

视频提“加密”但基础 set/get 并不自动实现它；Base64 不是加密，前端可见密钥也不消除 XSS 读取风险。偏好记录与刷新凭据应区分。

关联：[[前端面试 P096：令牌无感刷新、重试与安全边界]]。
