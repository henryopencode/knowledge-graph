package content

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	root := t.TempDir()
	s, err := Open(filepath.Join(root, "documents"), filepath.Join(root, "content.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func writeTest(t *testing.T, s *Store, id string, in WriteInput) Document {
	t.Helper()
	d, err := s.Write(context.Background(), id, in)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func TestDraftPublishWithdrawIsolation(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	d := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "公开标题", Content: "已审核正文", Slug: "public", Channels: &Channels{Blog: true, Graph: true}})
	posts, err := s.PublicPosts(ctx, "zh-CN", "")
	if err != nil || len(posts) != 0 {
		t.Fatal("draft was public", err)
	}
	if _, err = s.Publish(ctx, d.ID, d.Locale, d.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	d = writeTest(t, s, d.ID, WriteInput{Locale: d.Locale, Version: d.Version, Title: "内部新标题", Content: "SECRET_DRAFT_CONTENT", Slug: "public"})
	post, err := s.PublicPost(ctx, "zh-CN", "public")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(post)
	if strings.Contains(string(encoded), "SECRET") || post.Title != "公开标题" {
		t.Fatal("working version leaked", string(encoded))
	}
	if _, err = s.Write(ctx, d.ID, WriteInput{Locale: d.Locale, Version: "old", Title: "Conflict"}); ProblemOf(err).Code != "VERSION_CONFLICT" {
		t.Fatal("stale version accepted")
	}
	if err = s.Withdraw(ctx, d.ID, d.Locale, d.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	if _, err = s.PublicPost(ctx, "zh-CN", "public"); ProblemOf(err).Code != "DOCUMENT_NOT_FOUND" {
		t.Fatal("withdrawn document is public")
	}
}
func TestPrivateReferencesAndAssets(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	private := writeTest(t, s, "", WriteInput{Locale: "en", Title: "SECRET_TITLE", Content: "SECRET_BODY"})
	public := writeTest(t, s, "", WriteInput{Locale: "en", Title: "Public", Content: "A [[doc:" + private.ID + "|SECRET_ALIAS]] and [[SECRET_TITLE]].\n\n[private](../private.md)\n\n![chart](assets/chart.png)", Slug: "public", Channels: &Channels{Blog: true}})
	assets := filepath.Join(s.root, public.ID, "assets")
	if err := os.MkdirAll(assets, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assets, "chart.png"), []byte("old-image"), 0600); err != nil {
		t.Fatal(err)
	}
	post, err := s.Publish(ctx, public.ID, "en", public.PublicationVersion)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(post)
	if strings.Contains(string(raw), "SECRET") || strings.Contains(post.HTML, "private.md") {
		t.Fatal("private reference leaked", string(raw))
	}
	if err = os.WriteFile(filepath.Join(assets, "chart.png"), []byte("PRIVATE_NEW_IMAGE"), 0600); err != nil {
		t.Fatal(err)
	}
	asset, err := s.PublicAsset(ctx, public.ID, "en", "chart.png")
	if err != nil || string(asset.Data) != "old-image" {
		t.Fatal("asset snapshot changed", err)
	}
	if _, err = s.PrivateAsset(public.ID, "../../content.db"); err == nil {
		t.Fatal("path traversal accepted")
	}
	if err = s.Withdraw(ctx, public.ID, "en", public.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	if _, err = s.PublicAsset(ctx, public.ID, "en", "chart.png"); err == nil {
		t.Fatal("withdrawn attachment still public")
	}
}
func TestLanguagesAndIndependentChannels(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	base := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "知识", Content: "原文", Channels: &Channels{Graph: true}})
	writeTest(t, s, base.ID, WriteInput{Locale: "zh-TW", Title: "知識", Content: "譯文", SourceLocale: base.Locale, SourceRevision: base.Version, Channels: &Channels{Graph: true}})
	writeTest(t, s, base.ID, WriteInput{Locale: "en", Title: "Knowledge", Content: "Translation", Channels: &Channels{Graph: true}})
	for _, locale := range Locales {
		documents, err := s.List(locale, true)
		if err != nil || len(documents) != 1 || documents[0].Locale != locale {
			t.Fatalf("duplicated/missing graph node: %s %+v %v", locale, documents, err)
		}
	}
	blog := writeTest(t, s, "", WriteInput{Locale: "en", Title: "Blog only", Content: "Text", Channels: &Channels{Blog: true, Graph: false}})
	if _, err := s.Publish(ctx, blog.ID, "en", blog.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	graph, _ := s.List("en", true)
	if len(graph) != 1 {
		t.Fatal("blog-only node entered graph")
	}
	all, _ := s.List("en", false)
	if len(all) != 2 {
		t.Fatal("blog-only document missing from management")
	}
	writeTest(t, s, base.ID, WriteInput{Locale: "zh-CN", Title: "知识", Content: "新原文", Version: base.Version})
	translated, _ := s.List("zh-TW", false)
	for _, d := range translated {
		if d.ID == base.ID && !d.TranslationStale {
			t.Fatal("outdated translation not marked")
		}
	}
}
func TestLinksIgnoreCode(t *testing.T) {
	links := WikiLinks("[[real]] `[[inline]]`\n\n```go\n[[fenced]]\n```\n")
	if len(links) != 1 || links[0].Target != "real" {
		t.Fatalf("code became graph edges: %+v", links)
	}
}

func TestStoreRejectsSecondWriter(t *testing.T) {
	root := t.TempDir()
	documents := filepath.Join(root, "documents")
	database := filepath.Join(root, "index.db")
	first, err := Open(documents, database)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := Open(documents, database); err == nil {
		second.Close()
		first.Close()
		t.Fatal("two writer processes can open the store")
	}
	first.Close()
	third, err := Open(documents, database)
	if err != nil {
		t.Fatal("writer lock was not released", err)
	}
	third.Close()
}
