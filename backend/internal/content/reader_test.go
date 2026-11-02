package content

import (
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func directoryHashes(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	hashes := map[string][32]byte{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err == nil {
			hashes[path] = sha256.Sum256(data)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return hashes
}

func TestWorkspaceReaderAlongsideWriter(t *testing.T) {
	s := testStore(t)
	d := writeTest(t, s, "", WriteInput{Locale: "zh-CN", Title: "私有原文", Content: "未发布笔记", Channels: &Channels{Blog: true}})
	translated := writeTest(t, s, d.ID, WriteInput{Locale: "en", Title: "English note", Content: "English text"})
	linked := writeTest(t, s, "", WriteInput{Locale: "zh-TW", Title: "關聯筆記", Content: "[[doc:" + d.ID + "]]"})
	if _, err := s.Publish(context.Background(), d.ID, "zh-CN", translated.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	d = writeTest(t, s, d.ID, WriteInput{Locale: d.Locale, Version: d.Version, Title: d.Title, Content: "新的私有草稿"})
	before := directoryHashes(t, filepath.Dir(s.Root()))
	reader, err := OpenWorkspaceReader(s.Root())
	if err != nil {
		t.Fatal(err)
	}
	for _, locale := range Locales {
		documents, err := reader.List(locale)
		if err != nil || len(documents) != 2 {
			t.Fatalf("%s: %v, %v", locale, documents, err)
		}
		for _, item := range documents {
			if item.ID == d.ID {
				wantedLocale := "zh-CN"
				if locale == "en" {
					wantedLocale = "en"
				}
				if item.Locale != wantedLocale || len(item.AvailableLocales) != 2 || !reflect.DeepEqual(item.Incoming, []string{linked.ID}) || item.PublishedRevision != "" {
					t.Fatalf("incorrect projection: %+v", item)
				}
				if locale != "en" && item.Content != "新的私有草稿\n" {
					t.Fatalf("reader returned a stale snapshot: %q", item.Content)
				}
			}
		}
	}
	if after := directoryHashes(t, filepath.Dir(s.Root())); !reflect.DeepEqual(before, after) {
		t.Fatal("read-only access changed files or database")
	}
	post, err := s.PublicPost(context.Background(), "zh-CN", d.Slug)
	if err != nil || post.Title != d.Title || post.HTML == "" {
		t.Fatal("reader changed publication state", err)
	}
	d = writeTest(t, s, d.ID, WriteInput{Locale: d.Locale, Version: d.Version, Title: d.Title, Content: "再次更新"})
	documents, err := reader.List("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range documents {
		if item.ID == d.ID && item.Content != d.Content {
			t.Fatal("reader did not pick up subsequent writes")
		}
	}
}

func TestWorkspaceReaderFilesystemBoundary(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := OpenWorkspaceReader(missing); err == nil {
		t.Fatal("missing root accepted")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("reader created the missing root")
	}
	s := testStore(t)
	writeTest(t, s, "", WriteInput{Locale: "en", Title: "Note", Content: "text"})
	reader, err := OpenWorkspaceReader(s.Root())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = reader.List("fr"); ProblemOf(err).Code != "UNSUPPORTED_LOCALE" {
		t.Fatal("invalid locale accepted", err)
	}
	if err = os.Symlink(t.TempDir(), filepath.Join(s.Root(), "external")); err != nil {
		t.Fatal(err)
	}
	if _, err = reader.List("en"); err == nil {
		t.Fatal("symlink inside document root accepted")
	}
}
