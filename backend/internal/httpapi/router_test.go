package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBrowserOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(Config{AllowedOrigins: []string{"https://henryweb.site"}})
	for _, test := range []struct {
		name, method, origin string
		status               int
		allow                string
	}{
		{"same site", http.MethodGet, "https://henryweb.site", 200, "https://henryweb.site"},
		{"preflight", http.MethodOptions, "https://henryweb.site", 204, "https://henryweb.site"},
		{"untrusted site", http.MethodGet, "https://unrelated.example", 403, ""},
		{"health probe", http.MethodGet, "", 200, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/healthz", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != test.allow {
				t.Fatalf("allowed origin = %q, want %q", got, test.allow)
			}
		})
	}
}

func TestProjectChecksProtectedKnowledgeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	knowledge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer knowledge.Close()
	router := NewRouter(Config{KnowledgeURL: "https://knowledge.henryweb.site", KnowledgeInternalURL: knowledge.URL})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))
	var body struct {
		Projects []Project `json:"projects"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || len(body.Projects) != 1 || !body.Projects[0].Available {
		t.Fatalf("a login-protected knowledge service should count as reachable: %s", response.Body.String())
	}
	if body.Projects[0].URL != "https://knowledge.henryweb.site" {
		t.Fatal("internal upstream URL leaked into project link")
	}
}
