package content

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyMigrationPreservesBodyAndPrivacy(t *testing.T) {
	s := testStore(t)
	source := t.TempDir()
	path := filepath.Join(source, "legacy.md")
	raw := []byte("---\ntitle: Old note\nkind: concept\ntags: [go]\ncreated: 2026-10-01T00:00:00Z\ncustom:\n  owner: me\n---\n\nPrivate body with [[Old note]].\n")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	first, err := s.MigrateLegacy(ctx, source)
	if err != nil || len(first) != 1 {
		t.Fatal(first, err)
	}
	second, err := s.MigrateLegacy(ctx, source)
	if err != nil || len(second) != 1 || first[0].ID != second[0].ID {
		t.Fatal("migration is not idempotent", err)
	}
	d, err := s.Get(first[0].ID, "zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	if d.Content != "Private body with [[Old note]].\n" || d.Channels.Blog || !d.Channels.Graph || d.Properties["custom"] == nil {
		t.Fatalf("migration changed content or permissions: %+v", d)
	}
	original, _ := os.ReadFile(path)
	if string(original) != string(raw) {
		t.Fatal("migration changed the legacy source")
	}
	posts, err := s.PublicPosts(ctx, "zh-CN", "")
	if err != nil || len(posts) != 0 {
		t.Fatal("legacy notes became public", err)
	}
}
