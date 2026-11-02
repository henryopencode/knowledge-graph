"use strict";
const { t, locale } = WorkspaceI18n;

const $ = (selector, root = document) => root.querySelector(selector);
const $$ = (selector, root = document) => [...root.querySelectorAll(selector)];
const basePath = new URL(".", location.href).pathname;
const icons = {
  expand:
    '<path d="M8 3H3v5M16 3h5v5M3 16v5h5M21 16v5h-5M3 3l6 6M21 3l-6 6M3 21l6-6M21 21l-6-6"/>',
  back: '<path d="m10 5-7 7 7 7M3 12h18"/>',
  logout: '<path d="M9 4H4v16h5M14 7l5 5-5 5M9 12h10"/>',
  "panel-left":
    '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>',
  "panel-right":
    '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M15 4v16"/>',
  "chevron-down": '<path d="m6 9 6 6 6-6"/>',
  "folder-open": '<path d="M3 7V5h6l2 2h10v12H3V7Zm0 3h18"/>',
  folder: '<path d="M3 6h6l2 2h10v12H3V6Z"/>',
  "file-plus": '<path d="M13 3H5v18h14V9l-6-6ZM13 3v6h6M9 15h6M12 12v6"/>',
  files: '<path d="M8 3h8l4 4v12H8V3ZM16 3v4h4M4 7v15h12"/>',
  help: '<circle cx="12" cy="12" r="9"/><path d="M9.5 8.5a2.5 2.5 0 0 1 5 0c0 2-2.5 2-2.5 4M12 16h.01"/>',
  moon: '<path d="M20.5 13.5A9 9 0 0 1 10.5 3.5a9 9 0 1 0 10 10Z"/>',
  vault:
    '<rect x="3" y="3" width="18" height="18" rx="3"/><circle cx="12" cy="12" r="4"/><path d="M12 8v8M8 12h8"/>',
  sliders:
    '<path d="M4 6h8m4 0h4M4 12h3m4 0h9M4 18h10m4 0h2"/><circle cx="14" cy="6" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="16" cy="18" r="2"/>',
  search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4 4"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  minus: '<path d="M5 12h14"/>',
  close: '<path d="m6 6 12 12M18 6 6 18"/>',
  graph:
    '<circle cx="6" cy="6" r="2.5"/><circle cx="18" cy="9" r="2.5"/><circle cx="10" cy="19" r="2.5"/><path d="m8 7 7.5 1.5M7 8.5l2 8M16.5 11l-5 6"/>',
  list: '<path d="M9 6h11M9 12h11M9 18h11M4 6h.01M4 12h.01M4 18h.01"/>',
  project:
    '<rect x="3" y="5" width="18" height="15" rx="2"/><path d="M8 5V3h8v2M3 10h18M10 10v3h4v-3"/>',
  incident: '<path d="m12 3 10 17H2L12 3ZM12 9v4M12 16h.01"/>',
  decision:
    '<path d="M12 3v5M12 8 5 13v7M12 8l7 5v7M2 17l3 3 3-3M16 17l3 3 3-3"/>',
  concept:
    '<path d="M9 18h6M9 21h6M8 14a6 6 0 1 1 8 0c-1 1-1 2-1 3H9c0-1 0-2-1-3Z"/>',
  book: '<path d="M12 6c-3-2-6-2-10-1v15c4-1 7-1 10 1 3-2 6-2 10-1V5c-4-1-7-1-10 1ZM12 6v15"/>',
  import: '<path d="M12 3v12m-5-5 5 5 5-5M4 15v6h16v-6"/>',
  download: '<path d="M12 3v11m-4-4 4 4 4-4M5 15v5h14v-5"/>',
  refresh:
    '<path d="M20 7v5h-5M4 17v-5h5M6 6a8 8 0 0 1 13 3M5 15a8 8 0 0 0 13 3"/>',
  fit: '<path d="M8 3H3v5M16 3h5v5M3 16v5h5M21 16v5h-5"/>',
  file: '<path d="M14 2H5v20h14V7l-5-5ZM14 2v5h5M8 12h8M8 16h6"/>',
  edit: '<path d="m15 4 5 5M4 20l5-1L21 7l-5-5L4 14l-1 7Z"/>',
  trash: '<path d="M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7M14 10v7"/>',
  arrow: '<path d="M5 12h14m-5-5 5 5-5 5"/>',
  link: '<path d="m10 14 4-4M8 16l-1 1a4 4 0 0 1-6-6l4-4a4 4 0 0 1 6 0M16 8l1-1a4 4 0 0 1 6 6l-4 4a4 4 0 0 1-6 0" transform="translate(1 0) scale(.91)"/>',
  backlink: '<path d="m9 7-5 5 5 5M4 12h10a6 6 0 0 0 6-6V3"/>',
  spark:
    '<path d="m12 3 2.5 6.5L21 12l-6.5 2.5L12 21l-2.5-6.5L3 12l6.5-2.5L12 3Z"/>',
};
const icon = (name) =>
  `<svg class="icon" viewBox="0 0 24 24" aria-hidden="true">${icons[name] || icons.file}</svg>`;
$$("[data-icon]").forEach((el) =>
  el.replaceWith(fragment(icon(el.dataset.icon))),
);
function fragment(html) {
  const template = document.createElement("template");
  template.innerHTML = html;
  return template.content;
}
const escape = (value) =>
  String(value).replace(
    /[&<>"']/g,
    (char) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        char
      ],
  );
const kinds = {
  project: {
    label: t("label_project_context"),
    short: t("text_project"),
    color: "#3e806b",
  },
  incident: {
    label: t("label_incident_review"),
    short: t("text_incident"),
    color: "#c68a56",
  },
  decision: {
    label: t("label_technical_decision"),
    short: t("text_decision"),
    color: "#9a82bb",
  },
  concept: {
    label: t("label_concept"),
    short: t("text_concept"),
    color: "#6b91b3",
  },
};
const state = {
  notes: [],
  templates: {},
  selected: null,
  category: "all",
  domain: "",
  tag: "",
  query: "",
  depth: 0,
  mode: "graph",
  sort: "updated",
  online: true,
  hovered: null,
  spacing: 210,
  storage: "local",
};
const byId = (id) => state.notes.find((note) => note.id === id);
const degree = (note) => new Set([...note.outgoing, ...note.incoming]).size;
const kindPill = (kind) =>
  `<span class="kind-pill kind-${kind}">${icon(kind)}${kinds[kind].label}</span>`;
let toastTimer,
  searchTimer,
  editing = null,
  confirmResolve = null,
  draftSaved = false,
  readerNoteId = null,
  importing = false;

function notify(message, error = false) {
  const element = $("#toast");
  element.textContent = message;
  element.classList.toggle("error", error);
  element.hidden = false;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (element.hidden = true), error ? 6500 : 3500);
}
async function api(path, options = {}) {
  const response = await fetch(
    basePath +
      path.replace(/^\//, "") +
      (path.includes("?") ? "&" : "?") +
      "locale=" +
      locale,
    {
      ...options,
      headers: {
        "Content-Type": "application/json",
        "X-DevGraph-Request": "1",
        ...options.headers,
      },
    },
  );
  const data = await response.json();
  if (response.status === 401) location.reload();
  if (!response.ok) {
    const error = new Error(
      data.error?.message ||
        data.error ||
        t("operation_failed", { p0: response.status }),
    );
    error.code = data.error?.code || data.code;
    throw error;
  }
  return data;
}
function syncStatus(online, message) {
  state.online = online;
  $("#sync-status").textContent = message;
  $(".statusbar .status-dot").classList.toggle("offline", !online);
}
async function load(preferred, quiet = false) {
  try {
    const data = await api("/api/notes");
    state.notes = data.notes;
    state.templates = data.templates;
    state.storage = data.storage || "local";
    $("#logout-button").hidden = !data.authenticated;
    $("#storage-path").title = data.vault;
    $("#storage-path").textContent =
      state.storage === "server"
        ? t("text_markdown_workspace")
        : t("label_markdown_workspace");
    $("#editor-dialog .save-hint").innerHTML = t("save_location", {
      p0: icon("file"),
      p1: state.storage === "server" ? t("text_server") : t("text_local"),
    });
    $("#refresh-button").title = t("reload_files", {
      p0: state.storage === "server" ? t("text_server") : t("text_local"),
    });
    const allowed = filtered();
    const next = preferred || state.selected;
    state.selected = allowed.some((note) => note.id === next)
      ? next
      : (allowed.find((note) => note.kind === "project") || allowed[0])?.id ||
        null;
    const samples = state.notes.filter((note) => note.sample).length;
    $("#sample-status").textContent = samples
      ? t("sample_count", { p0: samples })
      : "";
    syncStatus(
      true,
      t("sync_time", {
        p0: state.storage === "server" ? t("text_server") : t("text_local"),
        p1: new Date().toLocaleTimeString(locale, {
          hour: "2-digit",
          minute: "2-digit",
        }),
      }),
    );
    render();
    if (!quiet) return true;
  } catch (error) {
    syncStatus(
      false,
      state.storage === "server"
        ? t("text_connection_lost_refresh_to_retry")
        : t("text_workspace_unavailable_refresh_to_retry"),
    );
    notify(t("read_failed", { p0: error.message }), true);
    $("#graph-status").textContent = t(
      "text_connection_lost_refresh_to_retry_2",
    );
  }
}
function filtered() {
  const query = state.query.toLocaleLowerCase();
  return state.notes.filter(
    (note) =>
      (state.category === "all" || note.kind === state.category) &&
      (!state.domain || note.category === state.domain) &&
      (!state.tag || note.tags.includes(state.tag)) &&
      (!query ||
        `${note.title}\n${note.category || ""}\n${note.content}\n${note.tags.join(" ")}`
          .toLocaleLowerCase()
          .includes(query)),
  );
}
function changeFilter() {
  const notes = filtered();
  if (!notes.some((note) => note.id === state.selected))
    state.selected = notes[0]?.id || null;
  render();
}
function render() {
  const edges = new Set(
    state.notes.flatMap((note) =>
      note.outgoing.map((id) => [note.id, id].sort().join("\0")),
    ),
  );
  $("#note-count").textContent = state.notes.length;
  $("#edge-count").textContent = edges.size;
  const categories = [
    ["all", t("label_all_documents"), "graph"],
    ...Object.entries(kinds).map(([id, info]) => [id, info.label, id]),
  ];
  $("#category-nav").innerHTML = categories
    .map(
      ([id, label, symbol]) =>
        `<button class="nav-item ${state.category === id ? "active" : ""}" data-category="${id}" aria-pressed="${state.category === id}">${icon(symbol)}<span>${label}</span><span class="count">${id === "all" ? state.notes.length : state.notes.filter((n) => n.kind === id).length}</span></button>`,
    )
    .join("");
  const domains = [
    ...new Set(state.notes.map((note) => note.category).filter(Boolean)),
  ].sort((a, b) => a.localeCompare(b, locale));
  $("#category-nav").insertAdjacentHTML(
    "beforeend",
    `<label class="domain-filter">${t("category")}<select id="domain-filter" aria-label="${t("category")}"><option value="">${t("label_all_documents")}</option>${domains.map((domain) => `<option value="${escape(domain)}" ${state.domain === domain ? "selected" : ""}>${escape(domain)}</option>`).join("")}</select></label>`,
  );
  const tags = [...new Set(state.notes.flatMap((note) => note.tags))].sort(
    (a, b) => a.localeCompare(b, locale),
  );
  $("#tags").innerHTML = tags.length
    ? tags
        .map(
          (tag) =>
            `<button class="tag-filter ${state.tag === tag ? "active" : ""}" data-tag="${escape(tag)}" aria-pressed="${state.tag === tag}"># ${escape(tag)}</button>`,
        )
        .join("")
    : `<span class="relation-empty">${t("empty_tags")}</span>`;
  $("#clear-tags").hidden = !state.tag;
  const summary = [
    state.category === "all"
      ? t("label_all_documents")
      : kinds[state.category].label,
    state.domain,
    state.tag && `# ${state.tag}`,
    state.query && `“${state.query}”`,
  ]
    .filter(Boolean)
    .join(" / ");
  $("#filter-summary").textContent = summary;
  $("#page-title").textContent =
    state.category === "all"
      ? t("label_knowledge_graph")
      : kinds[state.category].label;
  renderTree();
  renderList();
  renderDetails();
  drawGraph();
}
function renderTree() {
  const closed = new Set(
    $$(".tree-folder:not([open])").map((folder) => folder.dataset.folder),
  );
  const names = {
    project: t("text_01_projects"),
    incident: t("text_02_incidents"),
    decision: t("text_03_decisions"),
    concept: t("text_04_concepts"),
  };
  $("#file-tree").innerHTML = Object.keys(kinds)
    .map((kind) => {
      const notes = state.notes
        .filter((note) => note.kind === kind)
        .sort((a, b) => a.title.localeCompare(b.title, locale));
      return `<details class="tree-folder" data-folder="${kind}" ${closed.has(kind) ? "" : "open"}><summary>${icon("chevron-down")}${icon("folder")}<span>${names[kind]}</span><button class="count tree-kind-filter ${state.category === kind ? "active" : ""}" data-kind-filter="${kind}" title="${t("filter_kind", { kind: kinds[kind].label })}" aria-label="${t("filter_kind", { kind: kinds[kind].label })}" aria-pressed="${state.category === kind}">${notes.length}</button></summary><div class="tree-files">${notes.map((note) => `<button class="tree-note ${note.id === state.selected ? "selected" : ""}" data-note="${escape(note.id)}" title="${escape(note.title)}${note.sample ? t("text_example") : ""}">${icon("file")}<span>${escape(note.title)}</span></button>`).join("") || `<span class="relation-empty">${t("empty_notes")}</span>`}</div></details>`;
    })
    .join("");
}
function openInspector(open = true) {
  $("#workspace").classList.toggle("inspector-open", open);
  $("#toggle-inspector").setAttribute("aria-pressed", open);
}
function selectNote(id) {
  if (!byId(id)) return;
  state.selected = id;
  openInspector();
  if (!filtered().some((note) => note.id === id)) {
    state.category = "all";
    state.domain = "";
    state.tag = "";
    state.query = "";
    $("#search").value = "";
    render();
  } else {
    renderTree();
    renderDetails();
    renderList();
    if (state.depth) drawGraph();
    else highlightGraph();
  }
}
function prettyDate(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? value
    : date.toLocaleDateString(locale, {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
      });
}
function plainPreview(content) {
  return content
    .replace(/```[\s\S]*?```/g, "")
    .replace(/^#{1,6}\s+.*$/gm, "")
    .replace(/\[\[([^\]]+)\]\]/g, "$1")
    .replace(/[\n\r]+/g, " ")
    .replace(/[*_`>#]/g, "")
    .trim()
    .slice(0, 150);
}
function renderList() {
  const notes = [...filtered()].sort((a, b) =>
    state.sort === "title"
      ? a.title.localeCompare(b.title, locale)
      : state.sort === "links"
        ? degree(b) - degree(a)
        : (Date.parse(b.updated) || 0) - (Date.parse(a.updated) || 0),
  );
  $("#list-count").textContent = t("document_count", { p0: notes.length });
  $("#note-list").innerHTML = notes.length
    ? notes
        .map(
          (note) =>
            `<button class="note-card ${note.id === state.selected ? "selected" : ""}" data-note="${escape(note.id)}"><div class="note-card-top">${kindPill(note.kind)}${note.sample ? `<span class="sample-label">${t("sample")}</span>` : ""}</div><h3>${escape(note.title)}</h3><p>${escape(plainPreview(note.content) || t("text_this_document_has_no_content_yet"))}</p><div class="note-card-bottom"><span>${escape(prettyDate(note.updated))}</span><span>${escape(
              note.tags
                .slice(0, 2)
                .map((tag) => `#${tag}`)
                .join(" "),
            )}</span><span>${icon("link")}${t("connection_count", { count: degree(note) })}</span></div></button>`,
        )
        .join("")
    : `<div class="list-empty">${t("empty_list")}</div>`;
}
function resolveLink(target) {
  const normalized = target
    .split("|")[0]
    .split("#")[0]
    .replace(/^doc:/, "")
    .replace(/\.md$/, "")
    .trim();
  const exact = byId(normalized);
  if (exact) return exact;
  const candidates = state.notes.filter(
    (note) =>
      note.title === normalized || note.id.split("/").pop() === normalized,
  );
  return candidates.length === 1 ? candidates[0] : null;
}

// Raw HTML is always escaped. Only explicitly supported Markdown is rendered.
function inline(text) {
  const parts = [];
  const stash = (html) => `\u0000${parts.push(html) - 1}\u0000`;
  let safe = text
    .replace(/\u0000/g, "")
    .replace(/`([^`\n]+)`/g, (_, code) =>
      stash(`<code>${escape(code)}</code>`),
    );
  safe = safe.replace(/(?<!!)\[\[([^\]\n]+)\]\]/g, (_, link) => {
    const [target, alias] = link.split("|");
    const found = resolveLink(target);
    return stash(
      `<button class="wiki-link ${found ? "" : "missing"}" ${found ? `data-note="${escape(found.id)}"` : `data-missing="${escape(target.split("#")[0].replace(/\.md$/, "").trim())}"`} title="${found ? t("text_open_related_document") : t("text_document_missing_or_ambiguous_select_to_create")}">${escape(alias || target)}</button>`,
    );
  });
  safe = safe.replace(/\[([^\]\n]+)\]\(([^)\s]+)\)/g, (_, label, href) => {
    try {
      const url = new URL(href);
      if (!["https:", "http:", "mailto:"].includes(url.protocol))
        return escape(label);
      return stash(
        `<a href="${escape(url.href)}" target="_blank" rel="noopener noreferrer">${escape(label)}</a>`,
      );
    } catch {
      return escape(label);
    }
  });
  safe = escape(safe)
    .replace(/\*\*([^*\n]+)\*\*/g, "<strong>$1</strong>")
    .replace(/\*([^*\n]+)\*/g, "<em>$1</em>")
    .replace(/~~([^~\n]+)~~/g, "<del>$1</del>");
  return safe.replace(/\u0000(\d+)\u0000/g, (_, index) => parts[Number(index)]);
}
function markdown(content) {
  const lines = content.replace(/\r\n/g, "\n").split("\n");
  const result = [];
  let paragraph = [],
    listType = "",
    code = null,
    codeLines = [];
  function flushParagraph() {
    if (paragraph.length) {
      result.push(`<p>${paragraph.map(inline).join("<br>")}</p>`);
      paragraph = [];
    }
  }
  function closeList() {
    if (listType) {
      result.push(`</${listType}>`);
      listType = "";
    }
  }
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i],
      fence = line.match(/^\s*(`{3,}|~{3,})/);
    if (code) {
      if (fence && fence[1][0] === code[0] && fence[1].length >= code.length) {
        result.push(`<pre><code>${escape(codeLines.join("\n"))}</code></pre>`);
        code = null;
        codeLines = [];
      } else codeLines.push(line);
      continue;
    }
    if (fence) {
      flushParagraph();
      closeList();
      code = fence[1];
      continue;
    }
    if (!line.trim()) {
      flushParagraph();
      closeList();
      continue;
    }
    const heading = line.match(/^(#{1,6})\s+(.+)$/);
    if (heading) {
      flushParagraph();
      closeList();
      const level = Math.min(heading[1].length, 4);
      result.push(`<h${level}>${inline(heading[2])}</h${level}>`);
      continue;
    }
    if (/^\s*([-*_])(?:\s*\1){2,}\s*$/.test(line)) {
      flushParagraph();
      closeList();
      result.push("<hr>");
      continue;
    }
    if (/^>\s?/.test(line)) {
      flushParagraph();
      closeList();
      result.push(
        `<blockquote>${inline(line.replace(/^>\s?/, ""))}</blockquote>`,
      );
      continue;
    }
    const item = line.match(/^\s*(?:([-+*])|\d+\.)\s+(.+)$/);
    if (item) {
      flushParagraph();
      const nextType = item[1] ? "ul" : "ol";
      if (listType !== nextType) {
        closeList();
        result.push(`<${nextType}>`);
        listType = nextType;
      }
      result.push(`<li>${inline(item[2])}</li>`);
      continue;
    }
    if (
      line.includes("|") &&
      i + 1 < lines.length &&
      /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?\s*$/.test(lines[i + 1])
    ) {
      flushParagraph();
      closeList();
      const cells = (row) =>
        row
          .trim()
          .replace(/^\|/, "")
          .replace(/\|$/, "")
          .split("|")
          .map((x) => x.trim());
      result.push(
        `<table><thead><tr>${cells(line)
          .map((cell) => `<th>${inline(cell)}</th>`)
          .join("")}</tr></thead><tbody>`,
      );
      i++;
      while (
        i + 1 < lines.length &&
        lines[i + 1].includes("|") &&
        lines[i + 1].trim()
      )
        result.push(
          `<tr>${cells(lines[++i])
            .map((cell) => `<td>${inline(cell)}</td>`)
            .join("")}</tr>`,
        );
      result.push("</tbody></table>");
      continue;
    }
    closeList();
    paragraph.push(line);
  }
  flushParagraph();
  closeList();
  if (code)
    result.push(`<pre><code>${escape(codeLines.join("\n"))}</code></pre>`);
  return (
    result.join("") || `<p class="relation-empty">${t("empty_content")}</p>`
  );
}
function relationSection(ids, heading, symbol) {
  const notes = [...new Set(ids)].map(byId).filter(Boolean);
  return `<section class="relation-section"><div class="relation-heading">${icon(symbol)}${heading}<span class="relation-count">${notes.length}</span></div>${notes.length ? notes.map((note) => `<button class="relation-link" data-note="${escape(note.id)}"><span class="relation-dot dot-${note.kind}"></span>${escape(note.title)}${icon("arrow")}</button>`).join("") : `<p class="relation-empty">${t("empty_relations")}</p>`}</section>`;
}
function renderDetails() {
  const note = byId(state.selected);
  $("#fullscreen-note").disabled = !note;
  if (!note) {
    $("#note-panel").innerHTML =
      `<div class="detail-placeholder">${icon("book")}<h2>${t("empty_selection")}</h2><p>${state.notes.length ? t("text_no_matching_documents_try_another_keyword") : t("text_create_a_document_to_begin")}</p><button data-action="new" class="button button-primary">${t("label_new_document")}</button></div>`;
    if ($("#reader-dialog").open) $("#reader-dialog").close();
    return;
  }
  $("#note-panel").innerHTML =
    `<div class="detail-header"><div class="detail-meta"><div class="note-classification">${note.category ? `<span class="domain-pill">${escape(note.category)}</span>` : ""}${kindPill(note.kind)}</div><div class="detail-menu"><button class="icon-button" data-action="edit" title="${t("text_edit_document")}" aria-label="${t("text_edit_document")}">${icon("edit")}</button><button class="icon-button" data-action="delete" title="${t("text_move_to_trash")}" aria-label="${t("remove_document")}">${icon("trash")}</button></div></div><h2>${escape(note.title)}</h2><div class="detail-date">${t("updated_at", { date: escape(prettyDate(note.updated)) })}${note.sample ? `<span class="sample-label">${t("sample")}</span>` : ""}</div><div class="detail-tags">${note.tags.map((tag) => `<button data-tag="${escape(tag)}"># ${escape(tag)}</button>`).join("")}</div></div>${WorkspaceEditor.publicationControls(note)}<article class="detail-body">${note.html || markdown(note.content)}</article>${relationSection(note.outgoing, t("text_links_to"), "link")}${relationSection(note.incoming, t("text_referenced_by"), "backlink")}${note.missing.length ? `<section class="relation-section"><div class="relation-heading">${t("missing_documents")}<span class="relation-count">${note.missing.length}</span></div>${note.missing.map((title) => `<button class="relation-link" data-missing="${escape(title)}">${icon("plus")}${escape(title)}</button>`).join("")}</section>` : ""}${note.ambiguous.length ? `<section class="relation-section"><p class="missing-note">${t("ambiguous_links", { titles: note.ambiguous.map(escape).join(", ") })}</p></section>` : ""}<div class="file-path">${icon("file")} documents/${escape(note.id)}/${note.locale}.md</div>`;
  if ($("#reader-dialog").open) renderReader();
}

function renderReader() {
  const note = byId(state.selected);
  if (!note) return;
  const changed = readerNoteId !== note.id;
  readerNoteId = note.id;
  $("#reader-caption-title").textContent = note.title;
  $("#reader-back-label").textContent =
    state.mode === "list"
      ? t("text_back_to_documents")
      : t("label_back_to_graph");
  const tags = note.tags
    .map((tag) => `<span class="reader-tag"># ${escape(tag)}</span>`)
    .join("");
  $("#reader-content").innerHTML =
    `<div class="reader-metadata">${note.category ? `<span class="domain-pill">${escape(note.category)}</span>` : ""}${kindPill(note.kind)}<span>${t("updated_at", { date: escape(prettyDate(note.updated)) })}</span>${note.sample ? `<span class="sample-label">${t("sample")}</span>` : ""}</div><h1 id="reader-heading">${escape(note.title)}</h1>${tags ? `<div class="reader-tags">${tags}</div>` : ""}<article class="detail-body reader-body">${note.html || markdown(note.content)}</article><footer class="reader-relations">${note.outgoing.length ? relationSection(note.outgoing, t("text_related_documents"), "link") : ""}${note.incoming.length ? relationSection(note.incoming, t("text_documents_referencing_this_one"), "backlink") : ""}</footer>`;
  if (changed) $("#reader-scroll").scrollTop = 0;
}
function openReader() {
  if (!byId(state.selected)) return;
  renderReader();
  if (!$("#reader-dialog").open) $("#reader-dialog").showModal();
}
$("#fullscreen-note").onclick = openReader;
$("#close-reader").onclick = () => $("#reader-dialog").close();
$("#reader-dialog").addEventListener("close", () =>
  $("#fullscreen-note").focus(),
);
$("#reader-edit").onclick = () => {
  const note = byId(state.selected);
  if (note) openEditor(note);
};
$("#reader-content").onclick = (event) => {
  const link = event.target.closest("[data-note]");
  const missing = event.target.closest("[data-missing]");
  if (link) selectNote(link.dataset.note);
  else if (missing) openEditor(null, missing.dataset.missing.split("/").pop());
};

const svgNS = "http://www.w3.org/2000/svg";
const graph = $("#graph");
const positions = new Map();
let graphNodes = [],
  graphEdges = [],
  graphFrame = 0,
  width = 600,
  height = 480,
  view = { x: 0, y: 0, scale: 1 },
  drag = null;
function svgElement(name, attrs) {
  const node = document.createElementNS(svgNS, name);
  for (const [key, value] of Object.entries(attrs))
    node.setAttribute(key, value);
  return node;
}
function visibleGraphNotes() {
  const notes = filtered().filter((note) => note.channels.graph);
  if (!state.depth || !state.selected) return notes;
  const allowed = new Set(notes.map((note) => note.id)),
    visited = new Set([state.selected]);
  let frontier = [state.selected];
  for (let depth = 0; depth < state.depth; depth++) {
    const next = [];
    for (const id of frontier)
      for (const neighbor of [
        ...(byId(id)?.outgoing || []),
        ...(byId(id)?.incoming || []),
      ])
        if (allowed.has(neighbor) && !visited.has(neighbor)) {
          visited.add(neighbor);
          next.push(neighbor);
        }
    frontier = next;
  }
  return notes.filter((note) => visited.has(note.id));
}
function drawGraph() {
  cancelAnimationFrame(graphFrame);
  width = $("#graph-container").clientWidth || 600;
  height = $("#graph-container").clientHeight || 480;
  graph.setAttribute("viewBox", `0 0 ${width} ${height}`);
  const visible = visibleGraphNotes(),
    notes = visible.slice(0, 120);
  $("#graph-empty").hidden = notes.length !== 0;
  $("#graph-status").textContent =
    visible.length > 120
      ? t("graph_limit", { p0: visible.length })
      : state.depth
        ? t("graph_depth", { p0: notes.length, p1: state.depth })
        : t("graph_count", { p0: notes.length });
  $$("[data-depth]").forEach((button) => {
    const active = Number(button.dataset.depth) === state.depth;
    button.classList.toggle("active", active);
    button.setAttribute("aria-pressed", active);
  });
  const edgeRoot = $("#graph-edges"),
    nodeRoot = $("#graph-nodes");
  edgeRoot.replaceChildren();
  nodeRoot.replaceChildren();
  graphNodes = notes.map((note, index) => {
    const angle = index * 2.399963;
    const radius = index ? Math.sqrt(index / Math.max(1, notes.length - 1)) : 0;
    const previous = positions.get(note.id);
    const node = {
      ...note,
      x: previous?.x ?? width / 2 + Math.cos(angle) * radius * width * 0.35,
      y: previous?.y ?? height / 2 + Math.sin(angle) * radius * height * 0.36,
      vx: 0,
      vy: 0,
      radius: Math.min(36, 4 + degree(note) * 2.65),
      fixed: false,
    };
    const group = svgElement("g", {
      class: `graph-node node-${note.kind}${degree(note) >= 7 ? " hub" : ""}`,
      "data-id": note.id,
      tabindex: "0",
      role: "button",
      "aria-label": t("node_accessible", {
        p0: note.title,
        p1: kinds[note.kind].label,
        p2: degree(note),
      }),
    });
    group.append(
      svgElement("circle", {
        class: "node-halo",
        r: node.radius + 7,
        stroke: kinds[note.kind].color,
      }),
    );
    group.append(
      svgElement("circle", {
        class: "node-circle",
        r: node.radius,
        fill: kinds[note.kind].color,
      }),
    );
    const title = svgElement("title", {});
    title.textContent = note.title;
    group.append(title);
    const label = svgElement("text", {
      class: "node-label",
      y: node.radius + (degree(note) >= 7 ? 27 : 21),
      "text-anchor": "middle",
    });
    label.textContent =
      note.title.length > 16 ? note.title.slice(0, 15) + "…" : note.title;
    group.append(label);
    node.element = group;
    nodeRoot.append(group);
    return node;
  });
  const map = new Map(graphNodes.map((node) => [node.id, node])),
    keys = new Set();
  graphEdges = [];
  for (const note of graphNodes)
    for (const id of note.outgoing) {
      const target = map.get(id),
        key = [note.id, id].sort().join("\0");
      if (!target || keys.has(key)) continue;
      keys.add(key);
      const element = svgElement("line", { class: "graph-edge" });
      edgeRoot.append(element);
      graphEdges.push({ source: note, target, element });
    }
  highlightGraph();
  updateView();
  paintGraph();
  let ticks = 0;
  const animate = () => {
    for (let step = 0; step < 3; step++) simulate();
    paintGraph();
    if (++ticks < 85 && !matchMedia("(prefers-reduced-motion: reduce)").matches)
      graphFrame = requestAnimationFrame(animate);
    else {
      for (let step = 0; step < 80; step++) simulate();
      paintGraph();
    }
  };
  graphFrame = requestAnimationFrame(animate);
}
function simulate() {
  for (const node of graphNodes) {
    node.vx += (width / 2 - node.x) * 0.00035;
    node.vy += (height / 2 - node.y) * 0.00035;
  }
  for (let i = 0; i < graphNodes.length; i++)
    for (let j = i + 1; j < graphNodes.length; j++) {
      const a = graphNodes[i],
        b = graphNodes[j];
      let dx = b.x - a.x,
        dy = b.y - a.y;
      if (Math.abs(dx) + Math.abs(dy) < 0.1) {
        dx = 0.7;
        dy = 0.7;
      }
      const distance = Math.hypot(dx, dy),
        force = Math.min(4, 2600 / (distance * distance));
      a.vx -= (dx / distance) * force;
      a.vy -= (dy / distance) * force;
      b.vx += (dx / distance) * force;
      b.vy += (dy / distance) * force;
      const minX = Math.min(185, (a.title.length + b.title.length) * 4 + 28),
        minY = a.radius + b.radius + 48;
      if (Math.abs(dx) < minX && Math.abs(dy) < minY) {
        if (minX - Math.abs(dx) < minY - Math.abs(dy)) {
          const push = Math.sign(dx || 1) * (minX - Math.abs(dx)) * 0.035;
          a.vx -= push;
          b.vx += push;
        } else {
          const push = Math.sign(dy || 1) * (minY - Math.abs(dy)) * 0.04;
          a.vy -= push;
          b.vy += push;
        }
      }
    }
  for (const edge of graphEdges) {
    const dx = edge.target.x - edge.source.x,
      dy = edge.target.y - edge.source.y,
      distance = Math.hypot(dx, dy) || 1,
      force = (distance - state.spacing) * 0.003;
    edge.source.vx += (dx / distance) * force;
    edge.source.vy += (dy / distance) * force;
    edge.target.vx -= (dx / distance) * force;
    edge.target.vy -= (dy / distance) * force;
  }
  for (const node of graphNodes) {
    if (!node.fixed) {
      node.vx *= 0.79;
      node.vy *= 0.79;
      node.x += node.vx;
      node.y += node.vy;
    }
    const margin = Math.min(
      width / 3,
      Math.max(60, Math.min(node.title.length, 16) * 7 + 8),
    );
    node.x = Math.min(Math.max(node.x, margin), width - margin);
    node.y = Math.min(
      Math.max(node.y, node.radius + 65),
      height - node.radius - 75,
    );
  }
}
function paintGraph() {
  for (const node of graphNodes) {
    node.element.setAttribute(
      "transform",
      `translate(${node.x.toFixed(2)} ${node.y.toFixed(2)})`,
    );
    positions.set(node.id, { x: node.x, y: node.y });
  }
  for (const edge of graphEdges)
    for (const [key, value] of Object.entries({
      x1: edge.source.x,
      y1: edge.source.y,
      x2: edge.target.x,
      y2: edge.target.y,
    }))
      edge.element.setAttribute(key, value.toFixed(2));
}
function highlightGraph() {
  const hovered = byId(state.hovered),
    neighbors = new Set(
      hovered ? [...hovered.outgoing, ...hovered.incoming, hovered.id] : [],
    );
  for (const node of graphNodes) {
    node.element.classList.toggle("selected", node.id === state.selected);
    node.element.classList.toggle(
      "dim",
      Boolean(hovered) && !neighbors.has(node.id),
    );
    node.element.setAttribute("aria-pressed", node.id === state.selected);
  }
  for (const edge of graphEdges) {
    const connected =
      Boolean(hovered) &&
      (edge.source.id === state.hovered || edge.target.id === state.hovered);
    edge.element.classList.toggle("connected", connected);
    edge.element.classList.toggle("dim", Boolean(hovered) && !connected);
  }
}
graph.addEventListener("pointerover", (event) => {
  const id = event.target.closest("[data-id]")?.dataset.id || null;
  if (id !== state.hovered) {
    state.hovered = id;
    highlightGraph();
  }
});
graph.addEventListener("pointerleave", () => {
  state.hovered = null;
  highlightGraph();
});
function updateView() {
  $("#graph-world").setAttribute(
    "transform",
    `translate(${view.x} ${view.y}) scale(${view.scale})`,
  );
  $("#zoom-label").textContent = `${Math.round(view.scale * 100)}%`;
}
function zoom(factor, x = width / 2, y = height / 2) {
  const previous = view.scale,
    next = Math.max(0.35, Math.min(2.5, previous * factor));
  view.x = x - ((x - view.x) * next) / previous;
  view.y = y - ((y - view.y) * next) / previous;
  view.scale = next;
  updateView();
}
graph.addEventListener(
  "wheel",
  (event) => {
    event.preventDefault();
    const rect = graph.getBoundingClientRect();
    zoom(
      Math.exp(-event.deltaY * 0.001),
      event.clientX - rect.left,
      event.clientY - rect.top,
    );
  },
  { passive: false },
);
graph.addEventListener("pointerdown", (event) => {
  if (event.button !== 0) return;
  const element = event.target.closest("[data-id]"),
    node = element && graphNodes.find((n) => n.id === element.dataset.id);
  drag = {
    node,
    startX: event.clientX,
    startY: event.clientY,
    originalX: node?.x ?? view.x,
    originalY: node?.y ?? view.y,
    moved: false,
  };
  if (node) node.fixed = true;
  graph.setPointerCapture(event.pointerId);
});
graph.addEventListener("pointermove", (event) => {
  if (!drag) return;
  const dx = event.clientX - drag.startX,
    dy = event.clientY - drag.startY;
  if (Math.abs(dx) + Math.abs(dy) > 4) drag.moved = true;
  if (drag.node) {
    drag.node.x = drag.originalX + dx / view.scale;
    drag.node.y = drag.originalY + dy / view.scale;
    paintGraph();
  } else {
    view.x = drag.originalX + dx;
    view.y = drag.originalY + dy;
    updateView();
  }
});
function endDrag(event) {
  if (!drag) return;
  const current = drag;
  drag = null;
  if (current.node) {
    current.node.fixed = false;
    if (!current.moved && event.type !== "pointercancel")
      selectNote(current.node.id);
  }
  if (graph.hasPointerCapture(event.pointerId))
    graph.releasePointerCapture(event.pointerId);
}
graph.addEventListener("pointerup", endDrag);
graph.addEventListener("pointercancel", endDrag);
graph.addEventListener("keydown", (event) => {
  const id = event.target.closest("[data-id]")?.dataset.id;
  if (id && ["Enter", " "].includes(event.key)) {
    event.preventDefault();
    selectNote(id);
  }
});
$("#zoom-in").onclick = () => zoom(1.2);
$("#zoom-out").onclick = () => zoom(1 / 1.2);
$("#fit-graph").onclick = () => {
  view = { x: 0, y: 0, scale: 1 };
  positions.clear();
  drawGraph();
};
let resizeTimer;
new ResizeObserver(() => {
  clearTimeout(resizeTimer);
  resizeTimer = setTimeout(() => {
    if (state.mode === "graph") drawGraph();
  }, 100);
}).observe($("#graph-container"));

function switchMode(mode) {
  state.mode = mode;
  for (const name of ["graph", "list"]) {
    $(`#${name}-panel`).hidden = mode !== name;
    $(`#${name}-tab`).classList.toggle("active", mode === name);
    $(`#${name}-tab`).setAttribute("aria-selected", mode === name);
  }
  $("#ribbon-graph").classList.toggle("active", mode === "graph");
  $("#ribbon-list").classList.toggle("active", mode === "list");
  if (mode === "graph") drawGraph();
}
$("#graph-tab").onclick = () => switchMode("graph");
$("#list-tab").onclick = () => switchMode("list");
$("#category-nav").onclick = (event) => {
  const category = event.target.closest("[data-category]")?.dataset.category;
  if (category) {
    state.category = category;
    if (category === "all") state.domain = "";
    changeFilter();
  }
};
$("#category-nav").onchange = (event) => {
  if (event.target.id === "domain-filter") {
    state.domain = event.target.value;
    changeFilter();
  }
};
function setTag(tag) {
  state.tag = state.tag === tag ? "" : tag;
  changeFilter();
}
$("#tags").onclick = (event) => {
  const button = event.target.closest("[data-tag]");
  if (button) setTag(button.dataset.tag);
};
$("#clear-tags").onclick = () => {
  state.tag = "";
  changeFilter();
};
$("#search").addEventListener("input", (event) => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    state.query = event.target.value.trim();
    changeFilter();
  }, 100);
});
$("#sort-order").onchange = (event) => {
  state.sort = event.target.value;
  renderList();
};
$("#note-list").onclick = (event) => {
  const id = event.target.closest("[data-note]")?.dataset.note;
  if (id) selectNote(id);
};
$$("[data-depth]").forEach(
  (button) =>
    (button.onclick = () => {
      state.depth = Number(button.dataset.depth);
      drawGraph();
    }),
);
$("#refresh-button").onclick = async () => {
  if (await load(state.selected))
    notify(
      t("reloaded", {
        p0: state.storage === "server" ? t("text_server") : t("text_local"),
      }),
    );
};
$("#file-tree").onclick = (event) => {
  const category =
    event.target.closest("[data-kind-filter]")?.dataset.kindFilter;
  if (category) {
    event.preventDefault();
    state.category = state.category === category ? "all" : category;
    return changeFilter();
  }
  const id = event.target.closest("[data-note]")?.dataset.note;
  if (id) selectNote(id);
};
$("#toggle-sidebar").onclick = () => {
  if (matchMedia("(max-width: 540px)").matches)
    $("#workspace").classList.toggle("mobile-sidebar-open");
  else $("#workspace").classList.toggle("sidebar-collapsed");
};
$("#toggle-inspector").onclick = () =>
  openInspector(!$("#workspace").classList.contains("inspector-open"));
$("#close-inspector").onclick = () => openInspector(false);
$("#ribbon-graph").onclick = () => switchMode("graph");
$("#ribbon-list").onclick = () => switchMode("list");
$("#ribbon-search").onclick = () => $("#search").focus();
$("#logout-button").onclick = async () => {
  try {
    await api("/api/logout", { method: "POST", body: "{}" });
    location.reload();
  } catch (error) {
    notify(error.message, true);
  }
};
function inspectorMode(mode) {
  $("#inspector").classList.toggle("links-mode", mode === "links");
  for (const name of ["note", "links"]) {
    const active = mode === name;
    $(`#inspector-${name}-tab`).classList.toggle("active", active);
    $(`#inspector-${name}-tab`).setAttribute("aria-selected", active);
  }
}
$("#inspector-note-tab").onclick = () => inspectorMode("note");
$("#inspector-links-tab").onclick = () => inspectorMode("links");
$("#settings-button").onclick = () => {
  const panel = $("#graph-settings");
  panel.hidden = !panel.hidden;
  $("#settings-button").setAttribute("aria-expanded", !panel.hidden);
};
$("#colored-nodes").onchange = (event) =>
  $("#graph-panel").classList.toggle("colored", event.target.checked);
$("#show-edges").onchange = (event) =>
  $("#graph-panel").classList.toggle("hide-edges", !event.target.checked);
$("#node-spacing").oninput = (event) => {
  state.spacing = Number(event.target.value);
  positions.clear();
  drawGraph();
};
$("#theme-button").onclick = () => {
  const dark = document.documentElement.classList.toggle("dark");
  localStorage.setItem("devgraph:theme", dark ? "dark" : "light");
};
if (localStorage.getItem("devgraph:theme") === "dark")
  document.documentElement.classList.add("dark");
document.addEventListener("pointerdown", (event) => {
  if (!event.target.closest("#graph-settings, #settings-button")) {
    $("#graph-settings").hidden = true;
    $("#settings-button").setAttribute("aria-expanded", "false");
  }
});

function confirmAction(
  message,
  label = t("label_confirm"),
  title = t("label_confirm_action"),
) {
  $("#confirm-heading").textContent = title;
  $("#confirm-message").textContent = message;
  $("#confirm-ok").textContent = label;
  $("#confirm-dialog").showModal();
  return new Promise((resolve) => (confirmResolve = resolve));
}
function finishConfirm(answer) {
  $("#confirm-dialog").close();
  confirmResolve?.(answer);
  confirmResolve = null;
}
$("#confirm-ok").onclick = () => finishConfirm(true);
$("#confirm-cancel").onclick = () => finishConfirm(false);
$("#confirm-dialog").addEventListener("cancel", (event) => {
  event.preventDefault();
  finishConfirm(false);
});
function editorSnapshot() {
  return [
    $("#editor-title").value,
    $("#editor-kind").value,
    $("#editor-tags").value,
    $("#editor-content").value,
    JSON.stringify(WorkspaceEditor.metadata()),
  ].join("\0");
}
let initialSnapshot = "";
function openEditor(note = null, title = "") {
  editing = note;
  draftSaved = false;
  $("#editor-heading").textContent = note
    ? t("text_edit_document")
    : t("label_new_document");
  $("#editor-title").value = note?.title || title;
  $("#editor-title").readOnly = false;
  $("#editor-title").title = note
    ? t("text_document_ids_keep_links_stable_when_titles_change")
    : "";
  $("#editor-kind").value =
    note?.kind || (state.category === "all" ? "incident" : state.category);
  $("#editor-tags").value = note?.tags.join(", ") || state.tag;
  $("#editor-content").value =
    note?.content ?? state.templates[$("#editor-kind").value] ?? "";
  WorkspaceEditor.configure(note);
  if (!note && state.domain) $("#editor-category").value = state.domain;
  $("#editor-message").textContent = "";
  $("#link-suggestions").hidden = true;
  initialSnapshot = editorSnapshot();
  $("#editor-dialog").showModal();
  setTimeout(
    () => (note || title ? $("#editor-content") : $("#editor-title")).focus(),
    30,
  );
}
async function closeEditor() {
  if (
    editorSnapshot() !== initialSnapshot &&
    !draftSaved &&
    !(await confirmAction(
      t("text_you_have_unsaved_changes_close_without_saving"),
      t("text_discard_changes"),
      t("text_close_editor"),
    ))
  )
    return;
  $("#editor-dialog").close();
  editing = null;
}
$("#new-button").onclick = () => openEditor();
$("#empty-new-button").onclick = () => openEditor();
$("#tree-new-button").onclick = () => openEditor();
$("#tab-new-button").onclick = () => openEditor();
$("#close-editor").onclick = closeEditor;
$("#cancel-editor").onclick = closeEditor;
$("#editor-dialog").addEventListener("cancel", (event) => {
  event.preventDefault();
  closeEditor();
});
$("#editor-kind").onchange = () => {
  if (
    !editing &&
    Object.values(state.templates).includes($("#editor-content").value)
  )
    $("#editor-content").value = state.templates[$("#editor-kind").value];
};
$("#apply-template").onclick = async () => {
  if (
    $("#editor-content").value.trim() &&
    !(await confirmAction(
      t(
        "text_this_replaces_the_current_content_copy_anything_you_want_to_keep_",
      ),
      t("text_replace_with_template"),
      t("text_insert_template"),
    ))
  )
    return;
  $("#editor-content").value = state.templates[$("#editor-kind").value];
  updateSuggestions();
};
$("#editor-form").onsubmit = async (event) => {
  event.preventDefault();
  const button = $("#save-note");
  button.disabled = true;
  $("#editor-message").textContent = "";
  const payload = {
    ...WorkspaceEditor.metadata(),
    title: $("#editor-title").value.trim(),
    kind: $("#editor-kind").value,
    tags: $("#editor-tags")
      .value.split(/[,，]/)
      .map((tag) => tag.trim().replace(/^#/, ""))
      .filter(Boolean),
    content: $("#editor-content").value,
    ...(editing ? { version: editing.version } : {}),
  };
  try {
    const note = await api(
      editing ? `/api/notes/${encodeURIComponent(editing.id)}` : "/api/notes",
      { method: editing ? "PUT" : "POST", body: JSON.stringify(payload) },
    );
    draftSaved = true;
    $("#editor-dialog").close();
    editing = null;
    state.category = "all";
    state.domain = "";
    state.tag = "";
    state.query = "";
    $("#search").value = "";
    await load(note.id);
    openInspector();
    notify(
      t("saved", {
        p0: state.storage === "server" ? t("text_server") : t("text_local"),
      }),
    );
  } catch (error) {
    $("#editor-message").textContent = error.message;
  } finally {
    button.disabled = false;
  }
};
function updateSuggestions() {
  const input = $("#editor-content"),
    before = input.value.slice(0, input.selectionStart),
    match = before.match(/\[\[([^\]\n]*)$/);
  const area = $("#link-suggestions");
  if (!match) {
    area.hidden = true;
    return;
  }
  const candidates = state.notes
    .filter(
      (note) =>
        note.id !== editing?.id &&
        note.title.toLocaleLowerCase().includes(match[1].toLocaleLowerCase()),
    )
    .slice(0, 5);
  area.hidden = !candidates.length;
  area.innerHTML = candidates
    .map(
      (note) =>
        `<button type="button" data-insert="${escape(note.id)}">${kinds[note.kind].short} · ${escape(note.title)}</button>`,
    )
    .join("");
}
$("#editor-content").addEventListener("input", updateSuggestions);
$("#editor-content").addEventListener("click", updateSuggestions);
$("#link-suggestions").onclick = (event) => {
  const id = event.target.closest("[data-insert]")?.dataset.insert,
    note = byId(id);
  if (!note) return;
  const input = $("#editor-content"),
    start = input.selectionStart,
    before = input.value.slice(0, start),
    match = before.match(/\[\[([^\]\n]*)$/);
  if (!match) return;

  const replacement = `[[doc:${note.id}|${note.title}]]`;
  input.setRangeText(
    replacement,
    start - match[0].length,
    start + (input.value.slice(start, start + 2) === "]]" ? 2 : 0),
    "end",
  );
  input.focus();
  $("#link-suggestions").hidden = true;
};
$("#note-panel").onclick = async (event) => {
  const note = byId(state.selected),
    link = event.target.closest("[data-note]"),
    missing = event.target.closest("[data-missing]"),
    tag = event.target.closest("[data-tag]"),
    action = event.target.closest("[data-action]")?.dataset.action;
  if (link) return selectNote(link.dataset.note);
  if (missing)
    return openEditor(null, missing.dataset.missing.split("/").pop());
  if (tag) return setTag(tag.dataset.tag);
  if (action === "new") return openEditor();
  if (!note) return;
  if (action === "edit") openEditor(note);
  if (
    action === "delete" &&
    (await confirmAction(
      t("remove_confirm", { p0: note.title }),
      t("text_move_to_trash"),
      t("text_remove_this_document"),
    ))
  ) {
    try {
      await api(`/api/notes/${encodeURIComponent(note.id)}`, {
        method: "DELETE",
        body: JSON.stringify({ version: note.version }),
      });
      await load();
      notify(
        t(
          "text_moved_to_the_trash_directory_the_original_file_can_be_restored",
        ),
      );
    } catch (error) {
      notify(error.message, true);
    }
  }
};
function openImport() {
  if (importing) return;
  $("#import-message").textContent = "";
  $("#import-progress").textContent = "";
  $("#import-results").hidden = true;
  $("#import-results").replaceChildren();
  $("#import-done").textContent = t("label_close");
  $("#import-dialog").showModal();
}
function closeImport() {
  if (!importing) $("#import-dialog").close();
}
$("#import-button").onclick = openImport;
$("#open-import").onclick = openImport;
$("#close-import").onclick = closeImport;
$("#import-done").onclick = closeImport;
$("#import-dialog").addEventListener("cancel", (event) => {
  if (importing) event.preventDefault();
});
$("#import-dropzone").onclick = () => {
  if (!importing) $("#import-files").click();
};
$("#import-files").onchange = async (event) => {
  await importDocuments([...event.target.files]);
  event.target.value = "";
};
const importDropzone = $("#import-dropzone");
for (const type of ["dragenter", "dragover"])
  importDropzone.addEventListener(type, (event) => {
    event.preventDefault();
    if (!importing) {
      importDropzone.classList.add("drag-active");
      if (event.dataTransfer) event.dataTransfer.dropEffect = "copy";
    }
  });
importDropzone.addEventListener("dragleave", (event) => {
  if (!importDropzone.contains(event.relatedTarget))
    importDropzone.classList.remove("drag-active");
});
importDropzone.addEventListener("drop", (event) => {
  event.preventDefault();
  importDropzone.classList.remove("drag-active");
  if (!importing) importDocuments([...event.dataTransfer.files]);
});
async function importDocuments(files) {
  if (!files.length || importing) return;
  $("#import-message").textContent = "";
  $("#import-results").hidden = true;
  try {
    if (files.some((file) => !/\.(md|markdown|txt)$/i.test(file.name)))
      throw new Error(
        t(
          "text_supported_formats_md_markdown_txt_convert_pdf_or_word_files_first",
        ),
      );
    if (files.length > 50 || files.some((file) => file.size > 512000))
      throw new Error(
        t("text_import_up_to_50_files_each_no_larger_than_500_kb"),
      );
    if (files.reduce((total, file) => total + file.size, 0) > 2000000)
      throw new Error(
        t("text_this_batch_exceeds_2_mb_split_it_into_smaller_batches"),
      );
    importing = true;
    for (const id of ["import-dropzone", "close-import", "import-done"])
      $(`#${id}`).disabled = true;
    $("#import-progress").textContent = t("reading_files", {
      p0: files.length,
    });
    const payload = {
      files: await Promise.all(
        files.map(async (file) => ({
          name: file.name,
          content: await file.text(),
        })),
      ),
    };
    const body = JSON.stringify(payload);
    if (new Blob([body]).size > 2000000)
      throw new Error(
        t("text_the_request_exceeds_2_mb_split_it_into_smaller_batches"),
      );
    $("#import-progress").textContent = t("uploading_files", {
      p0: files.length,
    });
    const result = await api("/api/import", { method: "POST", body });
    state.category = "all";
    state.domain = "";
    state.tag = "";
    state.query = "";
    $("#search").value = "";
    await load(result.imported[0]);
    if (result.imported.length) openInspector();
    if (result.errors.length) {
      $("#import-progress").textContent = t("import_partial", {
        p0: result.imported.length,
        p1: result.errors.length,
      });
      $("#import-results").innerHTML =
        `<ul>${result.errors.map((item) => `<li><strong>${escape(item.name)}</strong><span>${escape(item.error)}</span></li>`).join("")}</ul>`;
      $("#import-results").hidden = false;
      $("#import-done").textContent = t("text_done");
    } else {
      $("#import-dialog").close();
      notify(
        t("imported", {
          p0: result.imported.length,
          p1:
            state.storage === "server"
              ? t("text_server")
              : t("text_local_workspace"),
        }),
      );
    }
  } catch (error) {
    $("#import-message").textContent = error.message;
    $("#import-progress").textContent = "";
  } finally {
    importing = false;
    for (const id of ["import-dropzone", "close-import", "import-done"])
      $(`#${id}`).disabled = false;
  }
}
$("#help-button").onclick = () => $("#help-dialog").showModal();
$("#close-help").onclick = () => $("#help-dialog").close();
document.addEventListener("keydown", (event) => {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
    event.preventDefault();
    if (!$$("dialog[open]").length) $("#search").focus();
  }
  if (
    (event.metaKey || event.ctrlKey) &&
    event.key === "Enter" &&
    $("#editor-dialog").open &&
    !$("#confirm-dialog").open
  ) {
    event.preventDefault();
    if (!$("#save-note").disabled) $("#editor-form").requestSubmit();
  }
  if (
    event.key === "?" &&
    !["INPUT", "TEXTAREA", "SELECT"].includes(document.activeElement.tagName) &&
    !$$("dialog[open]").length
  )
    $("#help-dialog").showModal();
});
window.addEventListener("beforeunload", (event) => {
  if (
    $("#editor-dialog").open &&
    editorSnapshot() !== initialSnapshot &&
    !draftSaved
  ) {
    event.preventDefault();
    event.returnValue = "";
  }
});
if (!navigator.platform.toLowerCase().includes("mac")) {
  $("#search-shortcut").textContent = "Ctrl K";
  $(".key-hint").textContent = "Ctrl ↵";
}
load();
