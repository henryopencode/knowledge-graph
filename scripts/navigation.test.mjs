import assert from "node:assert/strict";
import test from "node:test";
import "../web/navigation.js";

const { topics, compactTitle } = globalThis.WorkspaceNavigation;

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
