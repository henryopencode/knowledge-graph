package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/pbkdf2"
	"knowledgegraph/backend/internal/auth"
	"knowledgegraph/backend/internal/content"
)

func TestWorkspaceAuthenticationAndThreeLanguages(t *testing.T) {
	root := t.TempDir()
	store, err := content.Open(filepath.Join(root, "documents"), filepath.Join(root, "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	salt := []byte("0123456789abcdef")
	digest := pbkdf2.Key([]byte("test-password"), salt, 260000, 32, sha256.New)
	credentials, err := auth.New(store.DB(), auth.Config{Username: "owner", Salt: hex.EncodeToString(salt), PasswordHash: hex.EncodeToString(digest), Origins: []string{"https://knowledge.henryweb.site"}})
	if err != nil {
		t.Fatal(err)
	}
	router, err := NewWorkspace(WorkspaceConfig{Store: store, Auth: credentials, AssetsDir: filepath.Join("..", "..", "..", "web")})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ locale, title string }{{"zh-CN", "打开你的知识库"}, {"zh-TW", "開啟你的知識庫"}, {"en", "Open your knowledge workspace"}} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest("GET", "/?locale="+item.locale, nil))
		if response.Code != 200 || !strings.Contains(response.Body.String(), item.title) {
			t.Fatalf("login translation %s: %d %s", item.locale, response.Code, response.Body.String())
		}
	}
	request := httptest.NewRequest("GET", "/api/v1/private/documents?locale=en", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 401 || !strings.Contains(response.Body.String(), "AUTH_REQUIRED") {
		t.Fatal("private route is not protected", response.Body.String())
	}
	for _, origin := range []string{"https://unrelated.example", "https://knowledge.henryweb.site"} {
		payload, _ := json.Marshal(map[string]string{"username": "owner", "password": "test-password"})
		request = httptest.NewRequest("POST", "/api/login?locale=en", bytes.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", origin)
		request.Header.Set("X-DevGraph-Request", "1")
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if origin == "https://unrelated.example" {
			if response.Code != 403 {
				t.Fatal("cross-site login accepted")
			}
			continue
		}
		if response.Code != 200 {
			t.Fatal(response.Body.String())
		}
		cookies := response.Result().Cookies()
		if len(cookies) == 0 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteStrictMode {
			t.Fatal("session cookie protections missing")
		}
		request = httptest.NewRequest("GET", "/api/v1/private/documents?locale=en", nil)
		request.AddCookie(cookies[0])
		response = httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatal("authenticated private route unavailable", response.Body.String())
		}
		// Publishing and withdrawal require no language choice, including when
		// the interface locale has no matching Markdown variant.
		d, err := store.Write(context.Background(), "", content.WriteInput{Locale: "zh-CN", Title: "原文", Slug: "original", Content: "正文", Channels: &content.Channels{Blog: true, Graph: true}})
		if err != nil {
			t.Fatal(err)
		}
		public := NewRouter(Config{Store: store})
		for _, action := range []string{"preview", "publish", "withdraw"} {
			payload, _ := json.Marshal(map[string]string{"version": d.PublicationVersion})
			request = httptest.NewRequest("POST", "/api/v1/private/documents/"+d.ID+"/"+action+"?locale=en", bytes.NewReader(payload))
			request.AddCookie(cookies[0])
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Origin", "https://knowledge.henryweb.site")
			request.Header.Set("X-DevGraph-Request", "1")
			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != 200 {
				t.Fatalf("%s without locale: %d %s", action, response.Code, response.Body.String())
			}
			if action == "preview" {
				continue
			}
			for _, language := range content.Locales {
				response = httptest.NewRecorder()
				public.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/public/posts/original?locale="+language, nil))
				if action == "publish" && (response.Code != 200 || !strings.Contains(response.Body.String(), `"content_locale":"zh-CN"`)) {
					t.Fatal("publication omitted a language", language, response.Body.String())
				}
				if action == "withdraw" && response.Code != 404 {
					t.Fatal("withdrawal left a language public", language)
				}
			}
		}

	}
}
