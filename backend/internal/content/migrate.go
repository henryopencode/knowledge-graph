package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func LegacyMarkdown(raw []byte) (map[string]any, string, error) {
	normalized := strings.ReplaceAll(strings.TrimPrefix(string(raw), "\ufeff"), "\r\n", "\n")
	meta := map[string]any{}
	if !strings.HasPrefix(normalized, "---\n") {
		return meta, normalized, nil
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return nil, "", Fail("INVALID_DOCUMENT", 400)
	}
	if err := yaml.Unmarshal([]byte(normalized[4:4+end]), &meta); err != nil {
		return nil, "", err
	}
	return meta, strings.TrimLeft(normalized[4+end+5:], "\n"), nil
}

type MigrationEntry struct {
	LegacyID   string `json:"legacy_id"`
	ID         string `json:"id"`
	SHA256     string `json:"sha256"`
	BodySHA256 string `json:"body_sha256"`
}

func (s *Store) MigrateLegacy(ctx context.Context, source string) ([]MigrationEntry, error) {
	entries := []MigrationEntry{}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return Fail("INVALID_DOCUMENT", 400)
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		meta, body, err := LegacyMarkdown(raw)
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(source, path)
		legacy := strings.TrimSuffix(filepath.ToSlash(relative), ".md")
		id := uuid.NewSHA1(uuid.NameSpaceURL, []byte("henryweb:legacy:"+legacy)).String()
		title := stringValue(meta["title"])
		if title == "" {
			title = strings.TrimSuffix(entry.Name(), ".md")
		}
		kind := stringValue(meta["kind"])
		if !Kinds[kind] {
			kind = "concept"
		}
		created := legacyTime(meta["created"])
		updated := legacyTime(meta["updated"])
		tags := []string{}
		if values, ok := meta["tags"].([]any); ok {
			for _, value := range values {
				tags = append(tags, fmt.Sprint(value))
			}
		}
		sample, _ := meta["sample"].(bool)
		properties := map[string]any{}
		for key, value := range meta {
			switch key {
			case "title", "kind", "created", "updated", "tags", "sample":
			default:
				properties[key] = value
			}
		}
		d := Document{Metadata: Metadata{SchemaVersion: 1, ID: id, Locale: "zh-CN", Title: title, Slug: id, Kind: kind, Category: "notes", Tags: tags, Created: created, Updated: updated, LegacyID: legacy, Sample: sample, Channels: Channels{Graph: true}, Properties: properties}, Content: body}
		if err = s.AddDocument(ctx, d, false); err != nil {
			return err
		}
		migrated, err := s.Get(id, "zh-CN")
		if err != nil {
			return err
		}
		if strings.TrimSpace(migrated.Content) != strings.TrimSpace(body) {
			return fmt.Errorf("migration changed body for %s", legacy)
		}
		sum := sha256.Sum256(raw)
		bodySum := sha256.Sum256([]byte(strings.TrimSpace(body)))
		entries = append(entries, MigrationEntry{LegacyID: legacy, ID: id, SHA256: hex.EncodeToString(sum[:]), BodySHA256: hex.EncodeToString(bodySum[:])})
		return nil
	})
	return entries, err
}
func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
func legacyTime(value any) string {
	if stamp, ok := value.(time.Time); ok {
		return stamp.UTC().Format(time.RFC3339)
	}
	if parsed, err := time.Parse(time.RFC3339, stringValue(value)); err == nil {
		return parsed.UTC().Format(time.RFC3339)
	}
	return now()
}

func (s *Store) Seed(ctx context.Context, source string) error {
	documents := []Document{}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		d, err := Parse(raw)
		if err != nil {
			return err
		}
		documents = append(documents, d)
		return nil
	})
	if err != nil {
		return err
	}
	sort.SliceStable(documents, func(i, j int) bool { return documents[i].SourceLocale == "" && documents[j].SourceLocale != "" })
	for _, d := range documents {
		if d.SourceLocale != "" {
			source, err := s.Get(d.ID, d.SourceLocale)
			if err != nil {
				return err
			}
			d.SourceRevision = source.Version
		}
		if err = s.AddDocument(ctx, d, true); err != nil {
			return err
		}
	}
	return nil
}
