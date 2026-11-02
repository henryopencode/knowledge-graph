import assert from "node:assert/strict";
import test from "node:test";
import "../web/navigation.js";

const { topics, compactTitle, searchNotes, preview } = globalThis.WorkspaceNavigation;

test("course-wide tags do not hide useful topics or duplicate documents", () => {
  const notes = [
    { id: "guide", title: "Learning guide", category: "Frontend", kind: "project", tags: ["Interview", "JavaScript"] },
    ...["JavaScript", "Vue", "React"].flatMap((topic) =>
      Array.from({ length: 4 }, (_, i) => ({
        id: `${topic}-${i}`, title: `Lesson ${i}`, category: "Frontend", kind: "concept", tags: ["Interview", topic],
      })),
    ),
    { id: "private", title: "Personal note", category: "notes", kind: "concept", tags: [] },
  ];
  const before = structuredClone(notes);
  const result = topics(notes, "en");
  assert.deepEqual(new Set(result.map((group) => group.label)), new Set(["Frontend", "JavaScript", "Vue", "React", "notes"]));
  assert.equal(result[0].notes[0].id, "guide");
  assert.deepEqual(result.flatMap((group) => group.notes.map((note) => note.id)).sort(), notes.map((note) => note.id).sort());
  assert.deepEqual(notes, before, "navigation must not rewrite document metadata");
});

test("small and untagged workspaces remain browsable by their own categories", () => {
  const notes = [
    { id: "a", title: "Article", kind: "concept", category: "My category", tags: ["Only tag"] },
    { id: "b", title: "Another", kind: "incident", category: "", tags: [] },
  ];
  assert.deepEqual(new Set(topics(notes, "zh-TW").map((group) => group.label)), new Set(["My category", "notes"]));
  assert.deepEqual(topics([], "en"), []);
});

test("topic sorting keeps lesson numbers in order and ignores tag casing", () => {
  const notes = [10, 2, 1].map((number, i) => ({
    id: String(number), title: `Lesson ${number}`, kind: "concept", category: "Course", tags: [i ? "react" : "React"],
  }));
  notes.push({ id: "other", title: "Other", kind: "concept", category: "Misc", tags: [] });
  const result = topics(notes, "en");
  assert.equal(result.filter((group) => group.key === "tag:react").length, 1);
  assert.deepEqual(result[0].notes.map((note) => note.id), ["1", "2", "10"]);
});

test("compact course titles retain the lesson identity and actual subject", () => {
  assert.equal(compactTitle("前端面试 P004：浅拷贝与深拷贝"), "P004 · 浅拷贝与深拷贝");
  assert.equal(compactTitle("JavaScript 函数传参"), "JavaScript 函数传参");
  assert.equal(compactTitle("An arbitrary title: unchanged"), "An arbitrary title: unchanged");
});

test("search ranks a title above a body mention and searches every metadata field", () => {
  const notes = [
    { id: "body", title: "课程目录", tags: [], content: "包含 Vue 响应式的参考资料。", updated: "2026-11-02" },
    { id: "title", title: "Vue 响应式原理", tags: ["JavaScript"], content: "Proxy 拦截对象访问。", updated: "2026-10-01" },
    { id: "summary", title: "其他笔记", summary: "Vue 组件", category: "前端", tags: ["状态管理"], content: "正文", updated: "2026-10-20" },
  ];
  const before = structuredClone(notes);
  assert.deepEqual(searchNotes(notes, "VUE 响应式").map((note) => note.id), ["title", "body"]);
  assert.deepEqual(searchNotes(notes, "前端 状态管理").map((note) => note.id), ["summary"]);
  assert.deepEqual(searchNotes(notes, "vue 组件").map((note) => note.id), ["summary"]);
  assert.deepEqual(searchNotes(notes, "Vue 不存在"), []);
  assert.deepEqual(searchNotes(notes, "  ").map((note) => note.id), ["body", "summary", "title"]);
  assert.deepEqual(notes, before);
});

test("search previews show the matching passage without Markdown links or opaque IDs", () => {
  const body = "来源：[Vue 课程](https://example.com/long-link)\n\n" + "基础介绍。".repeat(90) + "关键结论：Proxy 会拦截对象的访问。[[doc:00000000-0000-0000-0000-000000000000|相关笔记]]";
  const snippet = preview(body, "Proxy");
  assert.ok(snippet.includes("Proxy 会拦截"));
  assert.ok(snippet.startsWith("…"));
  assert.ok(!snippet.includes("https://") && !snippet.includes("doc:"));
  assert.equal(preview("来源：[课程](https://example.com) [[doc:00000000-0000-0000-0000-000000000000|相关笔记]]"), "来源：课程 相关笔记");
  assert.ok(preview("```js\nconst value = 1;\n```").includes("const value = 1"));
});
