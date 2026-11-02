"use strict";

function configureEditor(note) {
  document.getElementById("editor-locale").value = note?.locale || locale;
  document.getElementById("editor-locale").disabled = false;
  document.getElementById("editor-summary").value = note?.summary || "";
  document.getElementById("editor-slug").value = note?.slug || "";
  document.getElementById("editor-category").value = note?.category || "notes";
  document.getElementById("editor-blog").checked =
    note?.channels?.blog || false;
  document.getElementById("editor-graph").checked =
    note?.channels?.graph ?? true;
  document.getElementById("review-translation").hidden = !note?.source_locale;
  syncBlogFields();
}

function editorMetadata() {
  return {
    locale: document.getElementById("editor-locale").value,
    publication_version: editing?.publication_version || "",
    summary: document.getElementById("editor-summary").value,
    slug: document.getElementById("editor-blog").checked
      ? document.getElementById("editor-slug").value
      : editing?.slug || "",
    category: document.getElementById("editor-category").value,
    channels: {
      blog: document.getElementById("editor-blog").checked,
      graph: document.getElementById("editor-graph").checked,
    },
    ...(editing?.source_locale
      ? {
          source_locale: editing.source_locale,
          source_revision: editing.source_revision,
        }
      : {}),
  };
}

function publicationControls(note) {
  const published = Boolean(note.published_revision);
  const changed = note.published_revision !== note.publication_version;
  const status = !note.channels.blog
    ? "visibility_private"
    : !published
      ? "draft"
      : changed
        ? "unpublished_changes"
        : "published";
  return `<section class="publication-controls"><div class="publication-heading"><span class="publication-status">${t(status)}</span></div>${note.locale !== locale ? `<p class="translation-status">${t("missing_translation")}</p>` : ""}${note.translation_stale ? `<p class="translation-status">${t("translation_stale")}</p>` : ""}<div class="publication-actions"><button class="button" data-action="edit">${t("publication_settings")}</button>${note.channels.blog ? `<button class="button ${changed ? "button-primary" : ""}" data-publish-action="preview">${t(changed ? "preview_and_publish" : "preview")}</button>` : ""}${published ? `<button class="text-button" data-publish-action="withdraw">${t("withdraw")}</button>` : ""}</div></section>`;
}

function syncBlogFields() {
  const enabled = document.getElementById("editor-blog").checked;
  document.getElementById("blog-fields").hidden = !enabled;
  for (const id of ["editor-summary", "editor-slug"])
    document.getElementById(id).disabled = !enabled;
  document
    .getElementById("editor-blog")
    .setAttribute("aria-expanded", String(enabled));
}
document
  .getElementById("editor-blog")
  .addEventListener("change", syncBlogFields);

let previewTarget = null;
const previewDialog = document.getElementById("preview-dialog");
const previewMessage = document.getElementById("preview-message");
const publishButton = document.getElementById("publish-preview");
function setPending(button, pending) {
  button.disabled = pending;
  button.setAttribute("aria-busy", String(pending));
}

async function updatePublication(target, action) {
  await api(`/api/v1/private/documents/${target.id}/${action}`, {
    method: "POST",
    body: JSON.stringify({ version: target.version }),
  });
  await load(target.id);
  notify(t(action === "publish" ? "publish_success" : "withdraw_success"));
}

document
  .getElementById("editor-locale")
  .addEventListener("change", async (event) => {
    const source = editing;
    const language = event.target.value;
    if (!source || language === source.locale) return;
    event.target.value = source.locale;
    if (
      editorSnapshot() !== initialSnapshot &&
      !(await confirmAction(
        t("text_you_have_unsaved_changes_close_without_saving"),
        t("text_discard_changes"),
      ))
    )
      return;
    event.target.disabled = true;
    try {
      const target = await api(
        `/api/v1/private/documents/${source.id}?locale=${language}`,
      );
      if (editing !== source) return;
      openEditor(target);
    } catch (error) {
      if (editing !== source) return;
      if (error.code !== "DOCUMENT_NOT_FOUND") {
        document.getElementById("editor-message").textContent = error.message;
      } else {
        const original = source.version ? source : byId(source.id);
        openEditor({
          ...source,
          locale: language,
          title: "",
          summary: "",
          content: "",
          version: "",
          source_locale: original.locale,
          source_revision: original.version,
          channels: {
            blog: source.channels.blog,
            graph: source.channels.graph,
          },
        });
      }
    } finally {
      event.target.disabled = false;
    }
  });

document
  .getElementById("note-panel")
  .addEventListener("click", async (event) => {
    const button = event.target.closest("[data-publish-action]");
    const note = byId(state.selected);
    if (!button || !note) return;
    const target = { id: note.id, version: note.publication_version };
    setPending(button, true);
    try {
      if (button.dataset.publishAction === "preview") {
        const post = await api(
          `/api/v1/private/documents/${target.id}/preview`,
          { method: "POST", body: JSON.stringify(target) },
        );
        previewTarget = target;
        document.getElementById("preview-title").textContent = post.title;
        document.getElementById("preview-body").innerHTML = post.html;
        publishButton.hidden =
          note.published_revision === note.publication_version;
        document.querySelector(".preview-footer").hidden = publishButton.hidden;
        document.getElementById("preview-notice").hidden = publishButton.hidden;
        previewMessage.textContent = "";
        previewDialog.showModal();
      } else if (
        button.dataset.publishAction === "withdraw" &&
        (await confirmAction(
          t("withdraw_confirm"),
          t("withdraw"),
          t("withdraw"),
        ))
      ) {
        await updatePublication(target, "withdraw");
      }
    } catch (error) {
      notify(error.message, true);
    } finally {
      setPending(button, false);
    }
  });

publishButton.addEventListener("click", async () => {
  if (!previewTarget) return;
  const target = previewTarget;
  setPending(publishButton, true);
  previewMessage.textContent = "";
  try {
    await updatePublication(target, "publish");
    previewDialog.close();
  } catch (error) {
    previewMessage.textContent = error.message;
  } finally {
    setPending(publishButton, false);
  }
});
document
  .getElementById("close-preview")
  .addEventListener("click", () => previewDialog.close());
previewDialog.addEventListener("close", () => {
  previewTarget = null;
});

document
  .getElementById("review-translation")
  .addEventListener("click", async () => {
    if (!editing?.source_locale) return;
    try {
      const source = await api(
        `/api/v1/private/documents/${editing.id}?locale=${editing.source_locale}`,
      );
      editing.source_revision = source.version;
      document.getElementById("editor-message").textContent = t(
        "translation_reviewed",
      );
    } catch (error) {
      document.getElementById("editor-message").textContent = error.message;
    }
  });

const workspaceLanguage = document.getElementById("workspace-language");
workspaceLanguage.value = locale;
workspaceLanguage.addEventListener("change", async () => {
  if (
    document.getElementById("editor-dialog").open &&
    editorSnapshot() !== initialSnapshot &&
    !(await confirmAction(
      t("text_you_have_unsaved_changes_close_without_saving"),
      t("text_discard_changes"),
    ))
  ) {
    workspaceLanguage.value = locale;
    return;
  }
  const url = new URL(location.href);
  url.searchParams.set("locale", workspaceLanguage.value);
  location.assign(url);
});
document.addEventListener("click", (event) => {
  const anchor = event.target.closest('a[href^="#doc:"]');
  if (anchor) {
    event.preventDefault();
    selectNote(anchor.getAttribute("href").slice(5));
  }
});
globalThis.WorkspaceEditor = {
  configure: configureEditor,
  metadata: editorMetadata,
  publicationControls,
};
