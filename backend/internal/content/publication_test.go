package content

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func currentDocument(t *testing.T, s *Store, id, locale string) Document {
	t.Helper()
	d, err := s.Get(id, locale)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestWholeDocumentPublicationAndFallback(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	base := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "原文", Content: "原文正文", Slug: "original", Channels: &Channels{Blog: true, Graph: true}})
	if _, err := s.Publish(ctx, base.ID, "en", base.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	for _, locale := range Locales {
		posts, err := s.PublicPosts(ctx, locale, "")
		if err != nil || len(posts) != 1 {
			t.Fatalf("%s: %+v %v", locale, posts, err)
		}
		p := posts[0]
		if p.ID != base.ID || p.Locale != locale || p.ContentLocale != "zh-CN" || len(p.AvailableLocales) != 1 || len(p.Translations) != 3 || !strings.Contains(p.HTML, "原文正文") {
			t.Fatalf("bad fallback: %+v", p)
		}
	}
	variants, _ := s.variants(base.ID)
	if len(variants) != 1 {
		t.Fatal("fallback created synthetic translations")
	}
	// A new translation cannot change shared visibility or become public on save.
	english := writeTest(t, s, base.ID, WriteInput{Locale: "en", Title: "Translation", Content: "UNPUBLISHED_ENGLISH", Slug: "translated", SourceLocale: base.Locale, SourceRevision: base.Version, Channels: &Channels{Blog: false, Graph: false}})
	if !english.Channels.Blog {
		t.Fatal("translation did not inherit document policy")
	}
	old, _ := s.PublicPost(ctx, "en", "original")
	if old.ContentLocale != "zh-CN" || strings.Contains(old.HTML, "UNPUBLISHED") {
		t.Fatal("draft translation leaked")
	}
	if _, err := s.Publish(ctx, base.ID, "zh-TW", base.PublicationVersion); ProblemOf(err).Code != "VERSION_CONFLICT" {
		t.Fatal("stale aggregate revision accepted", err)
	}
	if _, err := s.Publish(ctx, base.ID, "zh-TW", english.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	post, err := s.PublicPost(ctx, "en", "translated")
	if err != nil || post.ContentLocale != "en" || len(post.AvailableLocales) != 2 || post.Translations["zh-TW"] != "original" {
		t.Fatalf("actual translation unavailable: %+v %v", post, err)
	}
	// Global publication state is identical even though graph membership is not.
	zh := currentDocument(t, s, base.ID, "zh-CN")
	en := currentDocument(t, s, base.ID, "en")
	if zh.PublicationVersion != en.PublicationVersion || zh.PublishedRevision != en.PublishedRevision || zh.PublishedRevision != zh.PublicationVersion || !zh.Channels.Graph || en.Channels.Graph {
		t.Fatal("inconsistent group state")
	}
	// An edit in another language invalidates a preview's publication token.
	if _, err := s.Preview(ctx, base.ID, "zh-CN", en.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	changed := writeTest(t, s, base.ID, WriteInput{Locale: "en", Title: "Translation", Content: "NEW_PRIVATE_DRAFT", Version: en.Version})
	if _, err := s.Publish(ctx, base.ID, "zh-CN", en.PublicationVersion); ProblemOf(err).Code != "VERSION_CONFLICT" {
		t.Fatal("cross-language edit bypassed CAS", err)
	}
	if _, err := s.Preview(ctx, base.ID, "zh-TW", en.PublicationVersion); ProblemOf(err).Code != "VERSION_CONFLICT" {
		t.Fatal("stale preview accepted", err)
	}
	post, _ = s.PublicPost(ctx, "en", "translated")
	if strings.Contains(post.HTML, "NEW_PRIVATE") {
		t.Fatal("working edit replaced published translation")
	}
	// Withdrawal accepts any interface locale and revokes every route.
	if err := s.Withdraw(ctx, base.ID, "zh-TW", changed.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	for _, locale := range Locales {
		posts, _ := s.PublicPosts(ctx, locale, "")
		if len(posts) != 0 {
			t.Fatal("withdrawal left a language public", locale)
		}
	}
}

func TestSharedVisibilityAndPortablePolicy(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	base := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "Base", Content: "Body", Channels: &Channels{Blog: true, Graph: true}})
	en := writeTest(t, s, base.ID, WriteInput{Locale: "en", Title: "English", Content: "Body"})
	if _, err := s.Publish(ctx, base.ID, "en", en.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	disabled := writeTest(t, s, base.ID, WriteInput{Locale: "en", Version: en.Version, PublicationVersion: en.PublicationVersion, Title: "English", Content: "Body", Channels: &Channels{Blog: false, Graph: true}})
	zh := currentDocument(t, s, base.ID, "zh-CN")
	if zh.Channels.Blog || disabled.Channels.Blog || zh.PublishedRevision != "" {
		t.Fatal("visibility stayed language-specific")
	}
	for _, locale := range Locales {
		posts, _ := s.PublicPosts(ctx, locale, "")
		if len(posts) > 0 {
			t.Fatal("disabled language remains public", locale)
		}
	}
	if err := s.Reindex(ctx); err != nil {
		t.Fatal(err)
	}
	if currentDocument(t, s, base.ID, "zh-CN").Channels.Blog {
		t.Fatal("legacy frontmatter overrode shared policy")
	}
	// Saving a stale visibility change must not override another language's policy edit.
	if _, err := s.Write(ctx, base.ID, WriteInput{Locale: base.Locale, Version: base.Version, PublicationVersion: en.PublicationVersion, Title: "Base", Content: "Body", Channels: &Channels{Blog: true}}); ProblemOf(err).Code != "VERSION_CONFLICT" {
		t.Fatal("stale shared policy change accepted", err)
	}
	raw, err := s.Export()
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range archive.File {
		if strings.HasSuffix(f.Name, "/document.json") {
			found = true
		}
	}
	if !found {
		t.Fatal("export lost shared visibility policy")
	}
	// A deliberate shared-policy change while creating a translation is honored,
	// but enabling visibility alone still cannot republish withdrawn snapshots.
	tw := writeTest(t, s, base.ID, WriteInput{Locale: "zh-TW", Title: "繁體譯文", Content: "正文", PublicationVersion: disabled.PublicationVersion, Channels: &Channels{Blog: true}})
	if !tw.Channels.Blog || !currentDocument(t, s, base.ID, "zh-CN").Channels.Blog {
		t.Fatal("explicit shared visibility change was ignored")
	}
	if count, _ := s.PublishedCount(ctx); count != 0 {
		t.Fatal("enabling visibility resurrected snapshots")
	}
}

func TestFallbackLinksAndAssetRevocation(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	target := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "Target", Slug: "target", Content: "Target body", Channels: &Channels{Blog: true}})
	if _, err := s.Publish(ctx, target.ID, "zh-CN", target.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	private := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "SECRET_TARGET", Content: "SECRET"})
	d := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "Source", Slug: "source", Content: "[[Target]] [[doc:" + private.ID + "|SECRET_ALIAS]] ![image](assets/chart.png)", Channels: &Channels{Blog: true}})
	dir := filepath.Join(s.root, d.ID, "assets")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "chart.png"), []byte("published-image"), 0600)
	if _, err := s.Publish(ctx, d.ID, "zh-CN", d.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	for _, locale := range Locales {
		post, err := s.PublicPost(ctx, locale, "source")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(post.HTML, "/"+locale+"/posts/target") || strings.Contains(post.HTML, "SECRET") || !strings.Contains(post.HTML, "locale="+locale) {
			t.Fatal("bad projected links", locale, post.HTML)
		}
		asset, err := s.PublicAsset(ctx, d.ID, locale, "chart.png")
		if err != nil || string(asset.Data) != "published-image" {
			t.Fatal("fallback attachment unavailable", locale, err)
		}
	}
	if err := s.Withdraw(ctx, d.ID, "en", d.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	for _, locale := range Locales {
		if _, err := s.PublicAsset(ctx, d.ID, locale, "chart.png"); err == nil {
			t.Fatal("withdrawn fallback asset public", locale)
		}
	}
}

func TestProjectedSlugConflictIsAtomic(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := writeTest(t, s, "", WriteInput{Locale: "en", Title: "A", Slug: "same", Content: "A", Channels: &Channels{Blog: true}})
	if _, err := s.Publish(ctx, a.ID, "en", a.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	b := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "B", Slug: "same", Content: "B", Channels: &Channels{Blog: true}})
	if _, err := s.Publish(ctx, b.ID, "zh-CN", b.PublicationVersion); ProblemOf(err).Code != "SLUG_CONFLICT" {
		t.Fatal("fallback slug collision accepted", err)
	}
	for _, locale := range Locales {
		posts, _ := s.PublicPosts(ctx, locale, "")
		if len(posts) != 1 || posts[0].ID != a.ID {
			t.Fatal("conflict partially published", locale)
		}
	}
}

func TestBulkPublicationPreflightAndDraftDefaults(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "A", Slug: "same", Content: "A", Channels: &Channels{Graph: true}})
	b := writeTest(t, s, "", WriteInput{Locale: "en", Title: "B", Slug: "same", Content: "B"})
	if _, err := s.PublishAll(ctx); ProblemOf(err).Code != "SLUG_CONFLICT" {
		t.Fatal("bulk collision not caught", err)
	}
	if currentDocument(t, s, a.ID, a.Locale).Channels.Blog {
		t.Fatal("failed bulk changed policy")
	}
	writeTest(t, s, b.ID, WriteInput{Locale: b.Locale, Title: b.Title, Content: b.Content, Slug: "unique", Version: b.Version})
	results, err := s.PublishAll(ctx)
	if err != nil || len(results) != 2 {
		t.Fatal("bulk publish failed", results, err)
	}
	for _, locale := range Locales {
		posts, err := s.PublicPosts(ctx, locale, "")
		if err != nil || len(posts) != 2 {
			t.Fatal("bulk language set differs", locale, err)
		}
	}
	newDraft := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "New draft", Content: "PRIVATE"})
	if newDraft.Channels.Blog {
		t.Fatal("bulk operation changed defaults for new notes")
	}
	if err := s.Reindex(ctx); err != nil {
		t.Fatal(err)
	}
	count, err := s.PublishedCount(ctx)
	if err != nil || count != 2 {
		t.Fatal("reindex published later drafts", count, err)
	}
}

func TestLegacyPublicationMigrationPreservesSnapshots(t *testing.T) {
	root := t.TempDir()
	documents := filepath.Join(root, "documents")
	db := filepath.Join(root, "index.db")
	ctx := context.Background()
	s, err := Open(documents, db)
	if err != nil {
		t.Fatal(err)
	}
	base := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "Reviewed", Slug: "reviewed", Content: "Reviewed body", Channels: &Channels{Blog: true}})
	data, _ := json.Marshal(base)
	// Simulate a database from the previous per-language publisher.
	if _, err = s.db.Exec("INSERT INTO publications(id,locale,revision,data,published_at) VALUES(?,?,?,?,?)", base.ID, base.Locale, base.Version, string(data), now()); err != nil {
		t.Fatal(err)
	}
	writeTest(t, s, base.ID, WriteInput{Locale: base.Locale, Version: base.Version, Title: "PRIVATE_DRAFT", Slug: "reviewed", Content: "PRIVATE_DRAFT"})
	os.Remove(filepath.Join(documents, base.ID, "document.json"))
	s.Close()
	s, err = Open(documents, db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, locale := range Locales {
		p, err := s.PublicPost(ctx, locale, "reviewed")
		if err != nil || p.Title != "Reviewed" || strings.Contains(p.HTML, "PRIVATE") {
			t.Fatal("migration exposed draft", locale, p, err)
		}
	}
	current := currentDocument(t, s, base.ID, base.Locale)
	if current.PublishedRevision == "" || current.PublishedRevision == current.PublicationVersion {
		t.Fatal("migration lost draft status")
	}
	if err := s.Withdraw(ctx, base.ID, "en", current.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	if err := s.migratePublications(ctx); err != nil {
		t.Fatal(err)
	}
	count, _ := s.PublishedCount(ctx)
	if count != 0 {
		t.Fatal("migration resurrected withdrawal")
	}
}

func TestRemovingTranslationCannotCreateFallbackSlugCollision(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "A", Slug: "shared", Content: "A", Channels: &Channels{Blog: true}})
	ae := writeTest(t, s, a.ID, WriteInput{Locale: "en", Title: "A English", Slug: "a-english", Content: "A English", SourceLocale: a.Locale, SourceRevision: a.Version})
	if _, err := s.Publish(ctx, a.ID, "en", ae.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	b := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "B", Slug: "b-original", Content: "B", Channels: &Channels{Blog: true}})
	be := writeTest(t, s, b.ID, WriteInput{Locale: "en", Title: "B English", Slug: "shared", Content: "B English", SourceLocale: b.Locale, SourceRevision: b.Version})
	if _, err := s.Publish(ctx, b.ID, "en", be.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(ctx, a.ID, "en", ae.Version); ProblemOf(err).Code != "SLUG_CONFLICT" {
		t.Fatal("translation removal created ambiguous fallback route", err)
	}
	if _, err := s.PublicPost(ctx, "en", "a-english"); err != nil {
		t.Fatal("failed removal changed snapshot", err)
	}
	if _, err := s.Get(a.ID, "en"); err != nil {
		t.Fatal("failed removal deleted Markdown", err)
	}
}
