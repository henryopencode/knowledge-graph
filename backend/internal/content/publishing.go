package content

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Asset struct {
	Path string
	MIME string
	Data []byte
}
type published struct {
	Document Document
	At       string
}
type publicationPlan struct {
	Variants []Document
	Assets   map[string][]Asset
	Revision string
	At       string
}

func (s *Store) plan(variants []Document) (publicationPlan, error) {
	plan := publicationPlan{Variants: variants, Assets: map[string][]Asset{}, Revision: publicationVersion(variants), At: now()}
	for _, d := range variants {
		assets, err := s.collectAssets(d)
		if err != nil {
			return publicationPlan{}, err
		}
		plan.Assets[d.Locale] = assets
	}
	return plan, nil
}

// Every language snapshot and its attachments commit as one publication.
func (s *Store) commitPublications(ctx context.Context, plans []publicationPlan) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, plan := range plans {
		id := plan.Variants[0].ID
		for _, table := range []string{"publications", "publication_assets"} {
			if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id=?", id); err != nil {
				return err
			}
		}
		for _, d := range plan.Variants {
			data, err := json.Marshal(d)
			if err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO publications(id,locale,revision,data,published_at) VALUES(?,?,?,?,?)", id, d.Locale, d.Version, string(data), plan.At); err != nil {
				return err
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO publication_history(id,locale,revision,data,published_at) VALUES(?,?,?,?,?) ON CONFLICT DO NOTHING", id, d.Locale, d.Version, string(data), plan.At); err != nil {
				return err
			}
			for _, asset := range plan.Assets[d.Locale] {
				if _, err = tx.ExecContext(ctx, "INSERT INTO publication_assets(id,locale,path,mime,data) VALUES(?,?,?,?,?)", id, d.Locale, asset.Path, asset.MIME, asset.Data); err != nil {
					return err
				}
			}
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO publication_groups(id,revision,published_at) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET revision=excluded.revision,published_at=excluded.published_at", id, plan.Revision, plan.At); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) publicationDraft(id, locale, version string) ([]Document, error) {
	if !ValidLocale(locale) {
		return nil, Fail("UNSUPPORTED_LOCALE", 400)
	}
	variants, err := s.variants(id)
	if err != nil {
		return nil, err
	}
	if version == "" || version != publicationVersion(variants) {
		return nil, Fail("VERSION_CONFLICT", 409)
	}
	return variants, nil
}

func replacePublication(active []published, variants []Document, at string) []published {
	id := variants[0].ID
	result := []published{}
	for _, p := range active {
		if p.Document.ID != id {
			result = append(result, p)
		}
	}
	for _, d := range variants {
		result = append(result, published{Document: d, At: at})
	}
	return result
}

func validatePublicSlugs(active []published) error {
	for _, locale := range Locales {
		slugs := map[string]string{}
		for _, p := range projectPublications(active, locale) {
			d := p.Document
			if owner, exists := slugs[d.Slug]; exists && owner != d.ID {
				return Fail("SLUG_CONFLICT", 409)
			}
			slugs[d.Slug] = d.ID
		}
	}
	return nil
}

func (s *Store) Publish(ctx context.Context, id, locale, version string) (PublicPost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	variants, err := s.publicationDraft(id, locale, version)
	if err != nil {
		return PublicPost{}, err
	}
	if !variants[0].Channels.Blog {
		return PublicPost{}, Fail("BLOG_DISABLED", 409)
	}
	active, err := s.active(ctx)
	if err != nil {
		return PublicPost{}, err
	}
	plan, err := s.plan(variants)
	if err != nil {
		return PublicPost{}, err
	}
	active = replacePublication(active, variants, plan.At)
	if err = validatePublicSlugs(active); err != nil {
		return PublicPost{}, err
	}
	if err = s.commitPublications(ctx, []publicationPlan{plan}); err != nil {
		return PublicPost{}, err
	}
	for _, p := range projectPublications(active, locale) {
		if p.Document.ID == id {
			return s.toPublic(ctx, p, active, locale)
		}
	}
	return PublicPost{}, Fail("DOCUMENT_NOT_FOUND", 404)
}

func (s *Store) Withdraw(ctx context.Context, id, locale, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.publicationDraft(id, locale, version); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"publication_groups", "publications", "publication_assets"} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id=?", id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) publications(ctx context.Context, id string, grouped bool) ([]published, error) {
	query := "SELECT p.data,p.published_at FROM publications p"
	if grouped {
		query += " JOIN publication_groups g ON p.id=g.id"
	}
	args := []any{}
	if id != "" {
		query += " WHERE p.id=?"
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []published{}
	for rows.Next() {
		var raw, at string
		if err = rows.Scan(&raw, &at); err != nil {
			return nil, err
		}
		var d Document
		if err = json.Unmarshal([]byte(raw), &d); err != nil {
			return nil, err
		}
		items = append(items, published{Document: d, At: at})
	}
	return items, rows.Err()
}

func (s *Store) activeFor(ctx context.Context, id string) ([]published, error) {
	items, err := s.publications(ctx, id, true)
	if err != nil {
		return nil, err
	}
	result := []published{}
	for _, p := range items {
		current, err := s.read(p.Document.ID, p.Document.Locale)
		if err == nil && current.Channels.Blog {
			result = append(result, p)
		}
	}
	return result, nil
}
func (s *Store) active(ctx context.Context) ([]published, error) { return s.activeFor(ctx, "") }

// Select a real published translation, otherwise its original. Never read a draft
// as fallback and never create synthetic translation files.
func projectPublications(active []published, locale string) []published {
	groups := map[string][]published{}
	for _, p := range active {
		groups[p.Document.ID] = append(groups[p.Document.ID], p)
	}
	result := []published{}
	for _, variants := range groups {
		rank := map[string]int{"zh-CN": 0, "zh-TW": 1, "en": 2}
		sort.Slice(variants, func(i, j int) bool { return rank[variants[i].Document.Locale] < rank[variants[j].Document.Locale] })
		selected := 0
		for i, p := range variants {
			if p.Document.SourceLocale == "" {
				selected = i
				break
			}
		}
		for i, p := range variants {
			if p.Document.Locale == locale {
				selected = i
				break
			}
		}
		result = append(result, variants[selected])
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Document.ID < result[j].Document.ID })
	return result
}

func (s *Store) PublicPosts(ctx context.Context, locale, query string) ([]PublicPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !ValidLocale(locale) {
		return nil, Fail("UNSUPPORTED_LOCALE", 400)
	}
	active, err := s.active(ctx)
	if err != nil {
		return nil, err
	}
	posts := []PublicPost{}
	for _, p := range projectPublications(active, locale) {
		post, err := s.toPublic(ctx, p, active, locale)
		if err != nil {
			return nil, err
		}
		if query != "" && !strings.Contains(strings.ToLower(post.Title+" "+post.Description+" "+bluemonday.StrictPolicy().Sanitize(post.HTML)+" "+strings.Join(post.Tags, " ")), strings.ToLower(query)) {
			continue
		}
		posts = append(posts, post)
	}
	sort.Slice(posts, func(i, j int) bool {
		if posts[i].Date == posts[j].Date {
			return posts[i].Slug < posts[j].Slug
		}
		return posts[i].Date > posts[j].Date
	})
	return posts, nil
}

func (s *Store) PublicPost(ctx context.Context, locale, slug string) (PublicPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !ValidLocale(locale) {
		return PublicPost{}, Fail("UNSUPPORTED_LOCALE", 400)
	}
	active, err := s.active(ctx)
	if err != nil {
		return PublicPost{}, err
	}
	for _, p := range projectPublications(active, locale) {
		if p.Document.Slug == slug {
			return s.toPublic(ctx, p, active, locale)
		}
	}
	return PublicPost{}, Fail("DOCUMENT_NOT_FOUND", 404)
}

func publicLookup(active []published, locale string) map[string][]Document {
	selected := map[string]Document{}
	for _, p := range projectPublications(active, locale) {
		selected[p.Document.ID] = p.Document
	}
	lookup := map[string][]Document{}
	for _, p := range active {
		target := p.Document
		for _, key := range []string{target.ID, target.Title, target.LegacyID, filepath.Base(target.LegacyID)} {
			if key == "" || key == "." {
				continue
			}
			duplicate := false
			for _, d := range lookup[key] {
				duplicate = duplicate || d.ID == target.ID
			}
			if !duplicate {
				lookup[key] = append(lookup[key], selected[target.ID])
			}
		}
	}
	return lookup
}

func (s *Store) toPublic(ctx context.Context, p published, active []published, locale string) (PublicPost, error) {
	d := p.Document
	assets := map[string]bool{}
	rows, err := s.db.QueryContext(ctx, "SELECT path FROM publication_assets WHERE id=? AND locale=?", d.ID, d.Locale)
	if err != nil {
		return PublicPost{}, err
	}
	for rows.Next() {
		var path string
		if err = rows.Scan(&path); err != nil {
			rows.Close()
			return PublicPost{}, err
		}
		assets[path] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return PublicPost{}, err
	}
	translations := map[string]string{}
	available := []string{}
	for _, language := range Locales {
		for _, item := range projectPublications(active, language) {
			if item.Document.ID == d.ID {
				translations[language] = item.Document.Slug
			}
		}
		for _, item := range active {
			if item.Document.ID == d.ID && item.Document.Locale == language {
				available = append(available, language)
			}
		}
	}
	rendered := d
	rendered.Locale = locale // routes follow the interface; content_locale stays truthful.
	html, err := render(rendered, publicLookup(active, locale), assets, true)
	if err != nil {
		return PublicPost{}, err
	}
	minutes := max(1, (utf8.RuneCountInString(d.Content)+599)/600)
	if d.Locale == "en" {
		minutes = max(1, (len(strings.Fields(d.Content))+199)/200)
	}
	category := d.Category
	if category == "" {
		category = "notes"
	}
	return PublicPost{ID: d.ID, Locale: locale, ContentLocale: d.Locale, AvailableLocales: available, Title: d.Title, Description: d.Summary, Slug: d.Slug, Category: category, Tags: d.Tags, Date: d.Created[:10], ReadingMinutes: minutes, HTML: html, Revision: d.Version, Translations: translations}, nil
}

func neutral(locale string) string {
	switch locale {
	case "en":
		return "Related document"
	case "zh-TW":
		return "關聯文件"
	default:
		return "关联文档"
	}
}
func render(d Document, lookup map[string][]Document, assets map[string]bool, public bool) (string, error) {
	engine := markdownEngine()
	source := []byte(d.Content)
	tree := engine.Parser().Parse(text.NewReader(source))
	nodes := []ast.Node{}
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			nodes = append(nodes, n)
		}
		return ast.WalkContinue, nil
	})
	for _, node := range nodes {
		if link, ok := node.(*WikiLink); ok {
			targets := lookup[link.Target]
			if len(targets) == 1 {
				target := targets[0]
				label := link.Label
				if label == "" {
					label = target.Title
				}
				a := ast.NewLink()
				if public {
					a.Destination = []byte("/" + d.Locale + "/posts/" + target.Slug)
				} else {
					a.Destination = []byte("#doc:" + target.ID)
				}
				a.AppendChild(a, ast.NewString([]byte(label)))
				link.Parent().ReplaceChild(link.Parent(), link, a)
			} else {
				label := neutral(d.Locale)
				if !public {
					label = link.Target
				}
				link.Parent().ReplaceChild(link.Parent(), link, ast.NewString([]byte(label)))
			}
			continue
		}
		var destination *[]byte
		switch n := node.(type) {
		case *ast.Link:
			destination = &n.Destination
		case *ast.Image:
			destination = &n.Destination
		}
		if destination == nil {
			continue
		}
		raw := string(*destination)
		if strings.HasPrefix(raw, "#") {
			continue
		}
		parsed, err := url.Parse(raw)
		if err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http" || parsed.Scheme == "mailto") {
			continue
		}
		asset := strings.TrimPrefix(strings.TrimPrefix(raw, "./"), "assets/")
		if public && strings.HasPrefix(strings.TrimPrefix(raw, "./"), "assets/") && assets[asset] {
			*destination = []byte("/api/v1/public/documents/" + d.ID + "/assets/" + asset + "?locale=" + d.Locale)
			continue
		}
		if !public && strings.HasPrefix(strings.TrimPrefix(raw, "./"), "assets/") {
			*destination = []byte("./api/v1/private/documents/" + d.ID + "/assets/" + asset + "?locale=" + d.Locale)
			continue
		}
		if node.Parent() != nil {
			node.Parent().ReplaceChild(node.Parent(), node, ast.NewString([]byte(neutral(d.Locale))))
		}
	}
	var output bytes.Buffer
	if err := engine.Renderer().Render(&output, source, tree); err != nil {
		return "", err
	}
	return bluemonday.UGCPolicy().Sanitize(output.String()), nil
}
func (s *Store) RenderPrivate(d Document, documents []Document) (string, error) {
	lookup := map[string][]Document{}
	for _, target := range documents {
		for _, key := range []string{target.ID, target.Title, target.LegacyID} {
			if key != "" {
				lookup[key] = append(lookup[key], target)
			}
		}
	}
	return render(d, lookup, nil, false)
}

func (s *Store) collectAssets(d Document) ([]Asset, error) {
	source := []byte(d.Content)
	tree := markdownEngine().Parser().Parse(text.NewReader(source))
	paths := map[string]bool{}
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			var destination string
			switch node := n.(type) {
			case *ast.Image:
				destination = string(node.Destination)
			case *ast.Link:
				destination = string(node.Destination)
			}
			destination = strings.TrimPrefix(destination, "./")
			if strings.HasPrefix(destination, "assets/") {
				paths[strings.TrimPrefix(destination, "assets/")] = true
			}
		}
		return ast.WalkContinue, nil
	})
	result := []Asset{}
	total := 0
	for path := range paths {
		asset, err := s.asset(d.ID, path)
		if err != nil {
			return nil, err
		}
		total += len(asset.Data)
		if total > 20_000_000 {
			return nil, Fail("ASSET_TOO_LARGE", 400)
		}
		result = append(result, asset)
	}
	return result, nil
}
func (s *Store) asset(id, name string) (Asset, error) {
	if !ValidID(id) || name == "" || strings.ContainsAny(name, "\\\x00") || strings.HasPrefix(name, "/") {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	root := filepath.Join(s.root, id, "assets")
	path := filepath.Join(root, filepath.FromSlash(name))
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil || resolved != path {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	if info.Size() > 10_000_000 {
		return Asset{}, Fail("ASSET_TOO_LARGE", 400)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Asset{}, err
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if strings.HasPrefix(contentType, "text/html") || strings.Contains(contentType, "javascript") || strings.Contains(contentType, "svg") {
		contentType = "application/octet-stream"
	}
	return Asset{Path: name, MIME: contentType, Data: data}, nil
}
func (s *Store) PrivateAsset(id, name string) (Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.asset(id, name)
}
func (s *Store) PublicAsset(ctx context.Context, id, locale, name string) (Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !ValidLocale(locale) {
		return Asset{}, Fail("UNSUPPORTED_LOCALE", 400)
	}
	active, err := s.activeFor(ctx, id)
	if err != nil {
		return Asset{}, err
	}
	selected := projectPublications(active, locale)
	if len(selected) != 1 {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	var asset Asset
	asset.Path = name
	err = s.db.QueryRowContext(ctx, "SELECT mime,data FROM publication_assets WHERE id=? AND locale=? AND path=?", id, selected[0].Document.Locale, name).Scan(&asset.MIME, &asset.Data)
	if err != nil {
		return Asset{}, Fail("DOCUMENT_NOT_FOUND", 404)
	}
	return asset, nil
}

func (s *Store) Preview(ctx context.Context, id, locale, version string) (PublicPost, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	variants, err := s.publicationDraft(id, locale, version)
	if err != nil {
		return PublicPost{}, err
	}
	if !variants[0].Channels.Blog {
		return PublicPost{}, Fail("BLOG_DISABLED", 409)
	}
	active, err := s.active(ctx)
	if err != nil {
		return PublicPost{}, err
	}
	plan, err := s.plan(variants)
	if err != nil {
		return PublicPost{}, err
	}
	active = replacePublication(active, variants, "")
	if err = validatePublicSlugs(active); err != nil {
		return PublicPost{}, err
	}
	for _, p := range projectPublications(active, locale) {
		if p.Document.ID != id {
			continue
		}
		post, err := s.toPublic(ctx, p, active, locale)
		if err != nil {
			return PublicPost{}, err
		}
		paths := map[string]bool{}
		for _, asset := range plan.Assets[p.Document.Locale] {
			paths[asset.Path] = true
		}
		rendered := p.Document
		rendered.Locale = locale
		post.HTML, err = render(rendered, publicLookup(active, locale), paths, true)
		post.HTML = strings.ReplaceAll(post.HTML, "/api/v1/public/documents/"+id+"/assets/", "./api/v1/private/documents/"+id+"/assets/")
		return post, err
	}
	return PublicPost{}, Fail("DOCUMENT_NOT_FOUND", 404)
}

func (s *Store) AddDocument(ctx context.Context, d Document, publish bool) error {
	raw, err := Encode(d)
	if err != nil {
		return err
	}
	s.mu.Lock()
	path, err := s.path(d.ID, d.Locale)
	if err == nil {
		_, err = os.Stat(path)
		if err == nil {
			s.mu.Unlock()
			return nil
		}
		if !os.IsNotExist(err) {
			s.mu.Unlock()
			return err
		}
		err = atomicWrite(path, raw)
	}
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if err = s.Reindex(ctx); err != nil {
		return err
	}
	if publish {
		saved, err := s.Get(d.ID, d.Locale)
		if err != nil {
			return err
		}
		_, err = s.Publish(ctx, d.ID, d.Locale, saved.PublicationVersion)
		return err
	}
	return nil
}

func (s *Store) Export() ([]byte, error) { return s.exportFiles() }
func (s *Store) PublishedCount(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	posts, err := s.active(ctx)
	ids := map[string]bool{}
	for _, p := range posts {
		ids[p.Document.ID] = true
	}
	return len(ids), err
}
