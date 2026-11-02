package content

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
	_ "modernc.org/sqlite"
)

type Store struct {
	root       string
	db         *sql.DB
	mu         sync.RWMutex
	writerLock *os.File
}

func Open(root, database string) (*Store, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(root, ".writer.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		lock.Close()
		return nil, fmt.Errorf("content store already has an active writer: %w", err)
	}
	opened := false
	defer func() {
		if !opened {
			_ = unix.Flock(int(lock.Fd()), unix.LOCK_UN)
			_ = lock.Close()
		}
	}()
	if err := os.MkdirAll(filepath.Dir(database), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", database)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{root: root, db: db, writerLock: lock}
	_, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;
CREATE TABLE IF NOT EXISTS documents(id TEXT NOT NULL,locale TEXT NOT NULL,slug TEXT NOT NULL,version TEXT NOT NULL,data TEXT NOT NULL,PRIMARY KEY(id,locale));
CREATE TABLE IF NOT EXISTS publications(id TEXT NOT NULL,locale TEXT NOT NULL,revision TEXT NOT NULL,data TEXT NOT NULL,published_at TEXT NOT NULL,PRIMARY KEY(id,locale));
CREATE TABLE IF NOT EXISTS publication_groups(id TEXT PRIMARY KEY,revision TEXT NOT NULL,published_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS publication_history(id TEXT NOT NULL,locale TEXT NOT NULL,revision TEXT NOT NULL,data TEXT NOT NULL,published_at TEXT NOT NULL,PRIMARY KEY(id,locale,revision));
CREATE TABLE IF NOT EXISTS sessions(token_hash TEXT PRIMARY KEY,expires INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS publication_assets(id TEXT NOT NULL,locale TEXT NOT NULL,path TEXT NOT NULL,mime TEXT NOT NULL,data BLOB NOT NULL,PRIMARY KEY(id,locale,path));`)
	if err == nil {
		err = store.Reindex(context.Background())
	}
	if err == nil {
		err = store.migratePublications(context.Background())
	}
	if err != nil {
		db.Close()
		return nil, err
	}
	opened = true
	return store, nil
}
func (s *Store) Close() error {
	err := s.db.Close()
	_ = unix.Flock(int(s.writerLock.Fd()), unix.LOCK_UN)
	_ = s.writerLock.Close()
	return err
}
func (s *Store) DB() *sql.DB  { return s.db }
func (s *Store) Root() string { return s.root }

func (s *Store) path(id, locale string) (string, error) {
	if !ValidID(id) || !ValidLocale(locale) {
		return "", Fail("DOCUMENT_NOT_FOUND", 404)
	}
	path := filepath.Join(s.root, id, locale+".md")
	if info, err := os.Lstat(filepath.Dir(path)); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", Fail("INVALID_DOCUMENT", 400)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil && resolved != path {
		return "", Fail("INVALID_DOCUMENT", 400)
	}
	return path, nil
}
func (s *Store) read(id, locale string) (Document, error) {
	path, err := s.path(id, locale)
	if err != nil {
		return Document{}, err
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return Document{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	if err != nil {
		return Document{}, err
	}
	if !info.Mode().IsRegular() || info.Size() > 520_000 {
		return Document{}, Fail("INVALID_DOCUMENT", 400)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Document{}, err
	}
	d, err := Parse(raw)
	if err != nil {
		return Document{}, err
	}
	if d.ID != id || d.Locale != locale {
		return Document{}, Fail("INVALID_DOCUMENT", 400)
	}
	policy, err := s.policy(id)
	if err == nil {
		d.Channels.Blog = policy.Blog
	} else if !os.IsNotExist(err) {
		return Document{}, err
	}
	return d, nil
}
func (s *Store) readAll() ([]Document, error) {
	documents := []Document{}
	err := filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == s.root {
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in content store: %s", path)
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		relative, _ := filepath.Rel(s.root, path)
		parts := strings.Split(relative, string(filepath.Separator))
		if len(parts) != 2 {
			return fmt.Errorf("unexpected Markdown path: %s", relative)
		}
		d, err := s.read(parts[0], strings.TrimSuffix(parts[1], ".md"))
		if err != nil {
			return fmt.Errorf("%s: %w", relative, err)
		}
		documents = append(documents, d)
		return nil
	})
	return documents, err
}
func (s *Store) Reindex(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	documents, err := s.readAll()
	if err != nil {
		return err
	}
	for id, variants := range groupDocuments(documents) {
		if _, err = s.policy(id); os.IsNotExist(err) {
			blog := false
			for _, d := range variants {
				blog = blog || d.Channels.Blog
			}
			if err = s.writePolicy(id, blog); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	for i := range documents {
		policy, err := s.policy(documents[i].ID)
		if err != nil {
			return err
		}
		documents[i].Channels.Blog = policy.Blog
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "DELETE FROM documents"); err != nil {
		return err
	}
	for _, d := range documents {
		if err = saveIndex(ctx, tx, d); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func saveIndex(ctx context.Context, tx *sql.Tx, d Document) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO documents(id,locale,slug,version,data) VALUES(?,?,?,?,?) ON CONFLICT(id,locale) DO UPDATE SET slug=excluded.slug,version=excluded.version,data=excluded.data`, d.ID, d.Locale, d.Slug, d.Version, string(data))
	return err
}
func (s *Store) Get(id, locale string) (Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, err := s.read(id, locale)
	if err != nil {
		return Document{}, err
	}
	variants, err := s.variants(id)
	if err != nil {
		return Document{}, err
	}
	return s.decorate(d, variants), nil
}

// List chooses one translation per logical document. Missing translations remain explicit.
func (s *Store) List(locale string, graphOnly bool) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all, err := s.readAll()
	if err != nil {
		return nil, err
	}
	result := selectDocuments(all, locale, graphOnly)
	groups := groupDocuments(all)
	for i := range result {
		result[i] = s.decorate(result[i], groups[result[i].ID])
	}
	return result, nil
}

// selectDocuments is shared by the workspace and the filesystem-only reader.
func selectDocuments(all []Document, locale string, graphOnly bool) []Document {
	groups := groupDocuments(all)
	result := []Document{}
	for _, variants := range groups {
		sort.Slice(variants, func(i, j int) bool { return variants[i].Locale < variants[j].Locale })
		selected := -1
		for i, d := range variants {
			if graphOnly && !d.Channels.Graph {
				continue
			}
			if selected < 0 || d.Locale == "zh-CN" {
				selected = i
			}
			if d.Locale == locale {
				selected = i
				break
			}
		}
		if selected < 0 {
			continue
		}
		d := variants[selected]
		d.AvailableLocales = []string{}
		d.Outgoing = []string{}
		d.Incoming = []string{}
		d.Missing = []string{}
		d.Ambiguous = []string{}
		for _, variant := range variants {
			d.AvailableLocales = append(d.AvailableLocales, variant.Locale)
			if d.SourceLocale == variant.Locale && d.SourceRevision != "" {
				d.TranslationStale = d.SourceRevision != variant.Version
			}
		}
		result = append(result, d)
	}
	resolveRelations(result)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Updated == result[j].Updated {
			return result[i].ID < result[j].ID
		}
		return result[i].Updated > result[j].Updated
	})
	return result
}
func resolveRelations(documents []Document) {
	aliases := map[string][]int{}
	for i, d := range documents {
		for _, key := range []string{d.ID, d.Title, d.LegacyID, filepath.Base(d.LegacyID)} {
			if key != "" && key != "." {
				aliases[key] = appendUniqueInt(aliases[key], i)
			}
		}
	}
	for i := range documents {
		for _, link := range WikiLinks(documents[i].Content) {
			targets := aliases[link.Target]
			if len(targets) == 0 {
				documents[i].Missing = appendUnique(documents[i].Missing, link.Target)
			} else if len(targets) > 1 {
				documents[i].Ambiguous = appendUnique(documents[i].Ambiguous, link.Target)
			} else if targets[0] != i {
				j := targets[0]
				documents[i].Outgoing = appendUnique(documents[i].Outgoing, documents[j].ID)
				documents[j].Incoming = appendUnique(documents[j].Incoming, documents[i].ID)
			}
		}
	}
}
func appendUnique(values []string, value string) []string {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}
func appendUniqueInt(values []int, value int) []int {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}

func (s *Store) Write(ctx context.Context, id string, input WriteInput) (Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !ValidLocale(input.Locale) {
		return Document{}, Fail("UNSUPPORTED_LOCALE", 400)
	}
	isNew := id == ""
	if isNew {
		id = uuid.NewString()
	}
	previous, readErr := s.read(id, input.Locale)
	if readErr != nil && ProblemOf(readErr).Code != "DOCUMENT_NOT_FOUND" {
		return Document{}, readErr
	}
	if readErr == nil && input.Version != previous.Version {
		return Document{}, Fail("VERSION_CONFLICT", 409)
	}
	if readErr != nil && !isNew {
		variants, err := s.readAll()
		if err != nil {
			return Document{}, err
		}
		exists := false
		for _, d := range variants {
			if d.ID == id {
				exists = true
			}
		}
		if !exists {
			return Document{}, Fail("DOCUMENT_NOT_FOUND", 404)
		}
		if input.Version != "" {
			return Document{}, Fail("VERSION_CONFLICT", 409)
		}
	}
	stamp := now()
	d := Document{Metadata: Metadata{SchemaVersion: 1, ID: id, Locale: input.Locale, Title: strings.TrimSpace(input.Title), Summary: input.Summary, Slug: input.Slug, Kind: input.Kind, Tags: input.Tags, Channels: Channels{Graph: true}, Created: stamp, Updated: stamp, SourceLocale: input.SourceLocale, SourceRevision: input.SourceRevision}, Content: strings.ReplaceAll(input.Content, "\r\n", "\n")}
	d.Category = input.Category
	if readErr == nil {
		d.Created = previous.Created
		d.LegacyID = previous.LegacyID
		d.Sample = previous.Sample
		d.Properties = previous.Properties
		d.Channels = previous.Channels
		if d.Category == "" {
			d.Category = previous.Category
		}
		if d.Slug == "" {
			d.Slug = previous.Slug
		}
		if d.SourceLocale == "" {
			d.SourceLocale = previous.SourceLocale
			d.SourceRevision = previous.SourceRevision
		}
	}
	if d.Category == "" {
		d.Category = "notes"
	}
	if input.Channels != nil {
		d.Channels = *input.Channels
	}
	if !isNew {
		variants, err := s.variants(id)
		if err != nil {
			return Document{}, err
		}
		// New translations inherit visibility unless the caller explicitly changes
		// the shared policy using its aggregate revision.
		if input.Channels == nil || (readErr != nil && input.PublicationVersion == "") {
			d.Channels.Blog = variants[0].Channels.Blog
		} else if d.Channels.Blog != variants[0].Channels.Blog && input.PublicationVersion != "" && input.PublicationVersion != publicationVersion(variants) {
			return Document{}, Fail("VERSION_CONFLICT", 409)
		}
	}
	if d.Slug == "" {
		d.Slug = id
	}
	if d.Kind == "" {
		d.Kind = "concept"
	}
	if d.Tags == nil {
		d.Tags = []string{}
	}
	if d.SourceLocale != "" {
		source, err := s.read(id, d.SourceLocale)
		if err != nil {
			return Document{}, Fail("INVALID_TRANSLATION", 400)
		}
		if d.SourceRevision != source.Version && (readErr != nil || d.SourceRevision != previous.SourceRevision) {
			return Document{}, Fail("VERSION_CONFLICT", 409)
		}
	}
	raw, err := Encode(d)
	if err != nil {
		return Document{}, err
	}
	path, err := s.path(id, input.Locale)
	if err != nil {
		return Document{}, err
	}
	if err = atomicWrite(path, raw); err != nil {
		return Document{}, err
	}
	if err = s.writePolicy(id, d.Channels.Blog); err != nil {
		return Document{}, err
	}
	d, err = Parse(raw)
	if err != nil {
		return Document{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback()
	if err = saveIndex(ctx, tx, d); err != nil {
		return Document{}, err
	}
	if !d.Channels.Blog {
		for _, table := range []string{"publication_groups", "publications", "publication_assets"} {
			if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id=?", id); err != nil {
				return Document{}, err
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return Document{}, err
	}
	variants, err := s.variants(id)
	if err != nil {
		return Document{}, err
	}
	return s.decorate(d, variants), nil
}
func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".saving-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(data); err == nil {
		err = temporary.Sync()
	}
	closeErr := temporary.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temporary.Name(), path)
}
func (s *Store) Remove(ctx context.Context, id, locale, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, err := s.read(id, locale)
	if err != nil {
		return err
	}
	if d.Version != version {
		return Fail("VERSION_CONFLICT", 409)
	}
	active, err := s.active(ctx)
	if err != nil {
		return err
	}
	remaining := []published{}
	for _, p := range active {
		if p.Document.ID != id || p.Document.Locale != locale {
			remaining = append(remaining, p)
		}
	}
	if err = validatePublicSlugs(remaining); err != nil {
		return err
	}
	path, _ := s.path(id, locale)
	destination := filepath.Join(s.root, ".trash", id, strings.ReplaceAll(now(), ":", "-")+"-"+locale+".md")
	if err = os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	if err = os.Rename(path, destination); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"documents", "publications", "publication_assets"} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id=? AND locale=?", id, locale); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM publication_groups WHERE id=? AND NOT EXISTS(SELECT 1 FROM publications WHERE id=?)", id, id); err != nil {
		return err
	}
	return tx.Commit()
}
