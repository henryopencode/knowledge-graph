package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"knowledgegraph/backend/internal/auth"
	"knowledgegraph/backend/internal/content"
	"knowledgegraph/backend/internal/httpapi"
)

var version = "0.3.0"

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func server(address string, handler http.Handler) *http.Server {
	return &http.Server{Addr: address, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
}
func main() {
	store, err := content.Open(env("CONTENT_ROOT", "./var/documents"), env("CONTENT_DB", "./var/content.db"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	origins := strings.Split(env("WORKSPACE_ORIGINS", "https://knowledge.henryweb.site,https://101.34.211.31,http://127.0.0.1:8181,http://localhost:8181"), ",")
	credentials, err := auth.New(store.DB(), auth.Config{Username: env("DEVGRAPH_USERNAME", "devgraph"), PasswordHash: os.Getenv("DEVGRAPH_PASSWORD_HASH"), Salt: os.Getenv("DEVGRAPH_PASSWORD_SALT"), Secure: env("COOKIE_SECURE", "true") == "true", Origins: origins})
	if err != nil {
		log.Fatal(err)
	}
	workspace, err := httpapi.NewWorkspace(httpapi.WorkspaceConfig{Store: store, Auth: credentials, AssetsDir: env("KNOWLEDGE_ASSETS", "../web")})
	if err != nil {
		log.Fatal(err)
	}
	public := httpapi.NewRouter(httpapi.Config{Version: version, SiteURL: env("SITE_URL", "https://henryweb.site"), KnowledgeURL: env("KNOWLEDGE_URL", "https://knowledge.henryweb.site"), KnowledgeInternalURL: env("KNOWLEDGE_INTERNAL_URL", "http://127.0.0.1:8181"), AllowedOrigins: strings.Split(env("ALLOWED_ORIGINS", "https://henryweb.site,https://www.henryweb.site,http://127.0.0.1:3000,http://127.0.0.1:3210"), ","), Store: store})
	servers := []*http.Server{server(env("LISTEN_ADDR", "127.0.0.1:8180"), public), server(env("WORKSPACE_ADDR", "127.0.0.1:8181"), workspace)}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	failures := make(chan error, len(servers))
	for _, service := range servers {
		go func(s *http.Server) {
			log.Printf("listening=%s", s.Addr)
			if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				failures <- err
			}
		}(service)
	}
	select {
	case <-ctx.Done():
	case err := <-failures:
		log.Printf("server failed: %v", err)
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, service := range servers {
		if err := service.Shutdown(shutdown); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}
}
