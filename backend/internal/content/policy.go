package content

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Blog visibility belongs to the document, while Markdown and graph membership
// belong to language variants. Keep the shared policy portable with the Markdown.
type documentPolicy struct {
	SchemaVersion int  `json:"schema_version"`
	Blog          bool `json:"blog"`
}

func (s *Store) policyPath(id string) (string, error) {
	path, err := s.path(id, "zh-CN")
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), "document.json"), nil
}

func (s *Store) policy(id string) (documentPolicy, error) {
	path, err := s.policyPath(id)
	if err != nil {
		return documentPolicy{}, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return documentPolicy{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > 1024 {
		return documentPolicy{}, Fail("INVALID_DOCUMENT", 400)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return documentPolicy{}, err
	}
	var policy documentPolicy
	if err = json.Unmarshal(raw, &policy); err != nil || policy.SchemaVersion != 1 {
		return documentPolicy{}, Fail("INVALID_DOCUMENT", 400)
	}
	return policy, nil
}

func (s *Store) writePolicy(id string, blog bool) error {
	path, err := s.policyPath(id)
	if err != nil {
		return err
	}
	// Reject symlinks even though atomicWrite would replace rather than follow one.
	if _, err = s.policy(id); err != nil && !os.IsNotExist(err) {
		return err
	}
	raw, _ := json.MarshalIndent(documentPolicy{SchemaVersion: 1, Blog: blog}, "", "  ")
	return atomicWrite(path, append(raw, '\n'))
}

func groupDocuments(documents []Document) map[string][]Document {
	groups := map[string][]Document{}
	for _, d := range documents {
		groups[d.ID] = append(groups[d.ID], d)
	}
	return groups
}

// A preview becomes stale if any language changes. Raw Markdown versions remain
// separate for editing and translation-source tracking.
func publicationVersion(variants []Document) string {
	ordered := append([]Document(nil), variants...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Locale < ordered[j].Locale })
	hash := sha256.New()
	for _, d := range ordered {
		fmt.Fprintf(hash, "%s\x00%s\x00%t\n", d.Locale, d.Version, d.Channels.Blog)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func (s *Store) variants(id string) ([]Document, error) {
	variants := []Document{}
	for _, locale := range Locales {
		d, err := s.read(id, locale)
		if err != nil {
			if ProblemOf(err).Code == "DOCUMENT_NOT_FOUND" {
				continue
			}
			return nil, err
		}
		variants = append(variants, d)
	}
	if len(variants) == 0 {
		return nil, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	return variants, nil
}

func (s *Store) decorate(d Document, variants []Document) Document {
	d.PublicationVersion = publicationVersion(variants)
	_ = s.db.QueryRow("SELECT revision,published_at FROM publication_groups WHERE id=?", d.ID).Scan(&d.PublishedRevision, &d.PublishedAt)
	return d
}

// Migrate visibility without copying draft content into a publication. Existing
// per-language snapshots form a group; missing snapshots use the original at read time.
func (s *Store) migratePublications(ctx context.Context) error {
	items, err := s.publications(ctx, "", false)
	if err != nil {
		return err
	}
	groups := map[string][]Document{}
	times := map[string]string{}
	for _, p := range items {
		groups[p.Document.ID] = append(groups[p.Document.ID], p.Document)
		if p.At > times[p.Document.ID] {
			times[p.Document.ID] = p.At
		}
	}
	for id, variants := range groups {
		if _, err = s.db.ExecContext(ctx, "INSERT INTO publication_groups(id,revision,published_at) VALUES(?,?,?) ON CONFLICT DO NOTHING", id, publicationVersion(variants), times[id]); err != nil {
			return err
		}
	}
	return nil
}
