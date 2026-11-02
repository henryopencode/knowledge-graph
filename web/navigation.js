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

  function searchNotes(notes, query) {
    const terms = [...new Set(query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean))];
    return notes.map((note) => {
      const fields = [note.title, note.tags.join(" "), note.summary || "", note.category || "", note.content].map((value) => value.toLocaleLowerCase());
      let score = 0;
      for (const term of terms) {
        const weights = [12, 8, 4, 2, 1];
        let matched = false;
        fields.forEach((field, index) => {
          if (field.includes(term)) {
            matched = true;
            score += weights[index];
          }
        });
        if (!matched) return null;
      }
      return { note, score };
    }).filter(Boolean).sort((a, b) =>
      b.score - a.score || (Date.parse(b.note.updated) || 0) - (Date.parse(a.note.updated) || 0) || a.note.id.localeCompare(b.note.id),
    ).map((hit) => hit.note);
  }

  function preview(content, query = "") {
    const text = content
      .replace(/^\s*```[^\n]*$/gm, "")
      .replace(/\[\[([^\]]+)\]\]/g, (_, value) => value.includes("|") ? value.split("|").slice(1).join("|") : value.startsWith("doc:") ? "" : value)
      .replace(/!?\[([^\]]*)\]\([^)]*\)/g, "$1")
      .replace(/https?:\/\/\S+/g, "")
      .replace(/^[\s]*#{1,6}\s+/gm, "")
      .replace(/[*_`>#|]/g, "")
      .replace(/\s+/g, " ").trim();
    const terms = query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean);
    const lower = text.toLocaleLowerCase();
    const matches = terms.map((term) => lower.indexOf(term)).filter((index) => index >= 0);
    const start = matches.length ? Math.max(0, Math.min(...matches) - 45) : 0;
    const end = Math.min(text.length, start + 190);
    return (start ? "…" : "") + text.slice(start, end) + (end < text.length ? "…" : "");
  }

  return { topics, compactTitle, searchNotes, preview };
})();
