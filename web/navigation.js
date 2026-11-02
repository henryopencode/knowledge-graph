"use strict";

// Navigation is a view of existing metadata; it never rewrites a document's
// category, tags or identity. Broad tags shared by almost every note are not
// useful topics, so use the next most common tag and fall back to the category.
globalThis.WorkspaceNavigation = (() => {
  function topics(notes, locale) {
    const tags = new Map();
    for (const note of notes) {
      for (const tag of new Set(note.tags.map((value) => value.toLowerCase()))) {
        tags.set(tag, (tags.get(tag) || 0) + 1);
      }
    }
    const groups = new Map();
    for (const note of notes) {
      const candidates = note.kind === "project" ? [] : note.tags.filter((tag) => {
        const count = tags.get(tag.toLowerCase());
        return count >= 3 && count < notes.length * 0.85;
      });
      candidates.sort((a, b) =>
        tags.get(b.toLowerCase()) - tags.get(a.toLowerCase()) ||
        a.localeCompare(b, locale),
      );
      const tag = candidates[0];
      const key = tag ? `tag:${tag.toLowerCase()}` : `category:${note.category || "notes"}`;
      if (!groups.has(key)) groups.set(key, { key, label: tag || note.category || "notes", notes: [] });
      groups.get(key).notes.push(note);
    }
    for (const group of groups.values()) {
      group.notes.sort((a, b) =>
        Number(b.kind === "project") - Number(a.kind === "project") ||
        a.title.localeCompare(b.title, locale, { numeric: true }),
      );
    }
    return [...groups.values()].sort((a, b) =>
      Number(b.notes.some((note) => note.kind === "project")) -
        Number(a.notes.some((note) => note.kind === "project")) ||
      b.notes.length - a.notes.length || a.label.localeCompare(b.label, locale),
    );
  }

  function compactTitle(title) {
    return title.replace(/^.*?\s+(P\d{3,})[：:]\s*/, "$1 · ");
  }

  return { topics, compactTitle };
})();
