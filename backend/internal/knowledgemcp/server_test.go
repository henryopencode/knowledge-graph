package knowledgemcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"knowledgegraph/backend/internal/content"
)

func fixture(t *testing.T) (*content.Store, *mcp.ClientSession, context.Context) {
	t.Helper()
	root := t.TempDir()
	store, err := content.Open(filepath.Join(root, "documents"), filepath.Join(root, "content.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	server, err := New(store.Root(), "https://knowledge.example/graph/")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "knowledge-test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return store, cs, ctx
}

func write(t *testing.T, store *content.Store, id string, in content.WriteInput) content.Document {
	t.Helper()
	d, err := store.Write(context.Background(), id, in)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func call[T any](t *testing.T, ctx context.Context, client *mcp.ClientSession, name string, args any) T {
	t.Helper()
	result, err := client.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil || result.IsError {
		t.Fatalf("%s: %v, %+v", name, err, result)
	}
	if len(result.Content) == 0 {
		t.Fatal("tool result has no model-visible text")
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var output T
	if err = json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	return output
}

func TestToolsSearchReadAndFreshness(t *testing.T) {
	store, client, ctx := fixture(t)
	tools, err := client.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 3 {
		t.Fatal("expected exactly three tools", tools, err)
	}
	for _, tool := range tools.Tools {
		if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
			t.Fatal("tool is not marked read-only", tool.Name)
		}
	}
	main := write(t, store, "", content.WriteInput{Locale: "zh-CN", Title: "部署故障", Summary: "项目复盘", Content: strings.Repeat("说明🙂", 120) + "连接超时\nPRIVATE_DRAFT", Tags: []string{"Neurix"}})
	write(t, store, "", content.WriteInput{Locale: "en", Title: "Other note", Content: "neurix 部署故障 连接超时"})
	write(t, store, main.ID, content.WriteInput{Locale: "en", Title: "Deployment incident", Content: "Translated text"})
	search := call[searchOutput](t, ctx, client, "search_documents", map[string]any{"query": "部署故障 NEURIX", "limit": 1})
	if search.Total != 2 || len(search.Documents) != 1 || search.Documents[0].ID != main.ID || search.NextOffset == nil {
		t.Fatalf("ranking, Unicode or pagination failed: %+v", search)
	}
	if search.Documents[0].ContentLocale != "zh-CN" || !strings.Contains(search.Documents[0].SourceURL, "/graph/api/v1/private/documents/"+main.ID+"?locale=zh-CN") {
		t.Fatal("missing source identity", search.Documents[0])
	}
	matched := call[searchOutput](t, ctx, client, "search_documents", map[string]any{"query": "连接超时"})
	for _, hit := range matched.Documents {
		if !utf8.ValidString(hit.Excerpt) || !strings.Contains(hit.Excerpt, "连接超时") {
			t.Fatal("excerpt missed the matching Chinese passage", hit.Excerpt)
		}
	}
	page2 := call[searchOutput](t, ctx, client, "search_documents", map[string]any{"query": "部署故障 neurix", "offset": *search.NextOffset, "limit": 1})
	if len(page2.Documents) != 1 || page2.Documents[0].ID == main.ID || page2.NextOffset != nil {
		t.Fatal("search continuation failed", page2)
	}
	result := call[readOutput](t, ctx, client, "read_document", map[string]any{"id": main.ID, "locale": "zh-TW", "limit": 7})
	if result.Locale != "zh-TW" || result.ContentLocale != "zh-CN" || result.ID != main.ID || len([]rune(result.Content)) != 7 || result.NextOffset == nil {
		t.Fatalf("fallback or content pagination failed: %+v", result)
	}
	joined := result.Content
	for result.NextOffset != nil {
		result = call[readOutput](t, ctx, client, "read_document", map[string]any{"id": main.ID, "offset": *result.NextOffset, "limit": 100})
		joined += result.Content
	}
	if joined != main.Content || !utf8.ValidString(joined) {
		t.Fatal("Unicode pagination lost content")
	}
	english := call[readOutput](t, ctx, client, "read_document", map[string]any{"id": main.ID, "locale": "en"})
	if english.ContentLocale != "en" || english.ID != main.ID || !strings.Contains(english.Content, "Translated text") {
		t.Fatal("translation identity changed", english)
	}
	main = write(t, store, main.ID, content.WriteInput{Locale: main.Locale, Version: main.Version, Title: main.Title, Content: "LATEST_PRIVATE_DRAFT"})
	latest := call[readOutput](t, ctx, client, "read_document", map[string]any{"id": main.ID})
	if latest.Content != main.Content || latest.Version != main.Version {
		t.Fatal("MCP cached stale notes", latest)
	}
	empty := call[searchOutput](t, ctx, client, "search_documents", map[string]any{"query": "never-matches-this"})
	if empty.Total != 0 || len(empty.Documents) != 0 {
		t.Fatal("false search result", empty)
	}
}

func TestRelationsAndInputBoundary(t *testing.T) {
	store, client, ctx := fixture(t)
	target := write(t, store, "", content.WriteInput{Locale: "zh-CN", Title: "Target", Content: "private"})
	main := write(t, store, "", content.WriteInput{Locale: "zh-CN", Title: "Main", Content: "[[doc:" + target.ID + "]] [[missing]]\n\n`[[code-only]]`"})
	write(t, store, target.ID, content.WriteInput{Locale: target.Locale, Version: target.Version, Title: target.Title, Content: "[[doc:" + main.ID + "]]"})
	backlink := write(t, store, "", content.WriteInput{Locale: "en", Title: "Backlink", Content: "[[Main]]"})
	result := call[relatedOutput](t, ctx, client, "related_documents", map[string]any{"id": main.ID, "limit": 1})
	if result.Total != 2 || result.NextOffset == nil || len(result.Missing) != 1 || result.Missing[0] != "missing" {
		t.Fatal("relations or code exclusion failed", result)
	}
	second := call[relatedOutput](t, ctx, client, "related_documents", map[string]any{"id": main.ID, "offset": *result.NextOffset})
	directions := map[string]string{}
	for _, r := range append(result.Documents, second.Documents...) {
		directions[r.ID] = r.Direction
	}
	if directions[target.ID] != "both" || directions[backlink.ID] != "incoming" {
		t.Fatal("incorrect relation direction", directions)
	}
	for _, test := range []struct {
		tool string
		args map[string]any
	}{
		{"search_documents", map[string]any{"locale": "fr"}},
		{"search_documents", map[string]any{"limit": 31}},
		{"search_documents", map[string]any{"offset": -1}},
		{"search_documents", map[string]any{"query": strings.Repeat("长", 1001)}},
		{"read_document", map[string]any{"id": "../../credentials"}},
		{"read_document", map[string]any{"id": main.ID, "limit": 20001}},
		{"read_document", map[string]any{"id": "00000000-0000-0000-0000-000000000000"}},
		{"related_documents", map[string]any{"id": "not-an-id"}},
		{"write_document", map[string]any{"id": main.ID, "content": "changed"}},
	} {
		response, err := client.CallTool(ctx, &mcp.CallToolParams{Name: test.tool, Arguments: test.args})
		if err == nil && !response.IsError {
			t.Fatalf("invalid or mutation call accepted: %+v", test)
		}
	}
	unchanged := call[readOutput](t, ctx, client, "read_document", map[string]any{"id": main.ID})
	if unchanged.Content != main.Content {
		t.Fatal("invalid calls changed content")
	}
}
