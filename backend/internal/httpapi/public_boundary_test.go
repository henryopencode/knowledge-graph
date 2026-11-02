package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"knowledgegraph/backend/internal/content"
)

func TestIndependentPublicListenerExposesOnlyPublishedContent(t *testing.T) {
	root := t.TempDir()
	store, err := content.Open(filepath.Join(root, "documents"), filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	public, err := store.Write(ctx, "", content.WriteInput{Locale: "zh-CN", Title: "Published", Slug: "published", Content: "PUBLIC_BODY", Channels: &content.Channels{Blog: true, Graph: true}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Publish(ctx, public.ID, "en", public.PublicationVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Write(ctx, public.ID, content.WriteInput{Locale: "zh-CN", Title: "Working draft", Content: "PRIVATE_REVISION", Version: public.Version}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Write(ctx, "", content.WriteInput{Locale: "zh-CN", Title: "Secret", Slug: "secret", Content: "PRIVATE_DOCUMENT"}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{Store: store})
	for _, path := range []string{"/api/v1/public/posts?locale=en", "/api/v1/public/posts/published?locale=en"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		body := response.Body.String()
		if response.Code != 200 || !strings.Contains(body, "PUBLIC_BODY") || strings.Contains(body, "PRIVATE_") {
			t.Fatalf("public snapshot boundary failed: %d %s", response.Code, body)
		}
		if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Set-Cookie") != "" {
			t.Fatal("public content must not be cached or create a workspace session")
		}
		if !strings.Contains(body, `"content_locale":"zh-CN"`) {
			t.Fatal("published original fallback changed")
		}
	}
	for _, request := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/public/posts/secret"},
		{http.MethodGet, "/api/v1/private/documents"},
		{http.MethodGet, "/api/notes"},
		{http.MethodGet, "/app.js"},
		{http.MethodPost, "/api/login"},
		{http.MethodPost, "/api/v1/public/posts"},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(request.method, request.path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s %s unexpectedly exposed: %d", request.method, request.path, response.Code)
		}
	}
}
