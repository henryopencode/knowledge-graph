package httpapi

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"knowledgegraph/backend/internal/content"
	"knowledgegraph/backend/internal/i18n"
)

type Config struct {
	Version, SiteURL, KnowledgeURL, KnowledgeInternalURL string
	AllowedOrigins                                       []string
	Store                                                *content.Store
}
type Project struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Tags        []string `json:"tags"`
	Available   bool     `json:"available"`
}

func NewRouter(config Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), language(), cors(config.AllowedOrigins))
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}
	health := func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "henryweb-api", "version": config.Version})
	}
	router.GET("/healthz", health)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"name": "Henryweb API", "version": config.Version, "locales": i18n.Supported, "endpoints": []string{"/healthz", "/api/v1/public/posts", "/api/v1/site", "/api/v1/projects"}})
	})
	v1 := router.Group("/api/v1")
	v1.GET("/health", health)
	v1.GET("/site", func(c *gin.Context) {
		c.JSON(200, gin.H{"name": "Henry's Log", "author": "Henry", "url": config.SiteURL, "icp": "鄂ICP备2026021103号-3", "knowledge_url": config.KnowledgeURL, "locale": locale(c), "locales": i18n.Supported})
	})
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	v1.GET("/projects", func(c *gin.Context) {
		available := false
		if req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, config.KnowledgeInternalURL+"/healthz", nil); err == nil {
			if response, err := client.Do(req); err == nil {
				available = response.StatusCode < 500
				response.Body.Close()
			}
		}
		c.JSON(200, gin.H{"projects": []Project{{ID: "knowledge", Name: i18n.Message(locale(c), "project_name"), Description: i18n.Message(locale(c), "project_description"), URL: config.KnowledgeURL, Tags: []string{"knowledge", "markdown", "visualization"}, Available: available}}})
	})
	public := v1.Group("/public")
	public.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		if config.Store == nil {
			respondError(c, content.Fail("INTERNAL_ERROR", 503))
			return
		}
		c.Next()
	})
	public.GET("/posts", func(c *gin.Context) {
		posts, err := config.Store.PublicPosts(c.Request.Context(), locale(c), c.Query("q"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(200, gin.H{"posts": posts, "locale": locale(c)})
	})
	public.GET("/posts/:slug", func(c *gin.Context) {
		post, err := config.Store.PublicPost(c.Request.Context(), locale(c), c.Param("slug"))
		if err != nil {
			respondError(c, err)
			return
		}
		c.JSON(200, post)
	})
	public.GET("/documents/:id/assets/*path", func(c *gin.Context) {
		asset, err := config.Store.PublicAsset(c.Request.Context(), c.Param("id"), locale(c), strings.TrimPrefix(c.Param("path"), "/"))
		if err != nil {
			respondError(c, err)
			return
		}
		serveAsset(c, asset)
	})
	router.NoRoute(func(c *gin.Context) { respondError(c, content.Fail("NOT_FOUND", 404)) })
	return router
}

func locale(c *gin.Context) string { return c.GetString("locale") }
func language() gin.HandlerFunc {
	return func(c *gin.Context) {
		selected := i18n.FromRequest(c.Request)
		c.Set("locale", selected)
		c.Header("Content-Language", selected)
		c.Header("Vary", "Accept-Language, Cookie, Origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Request-ID", uuid.NewString())
		if requested := c.Query("locale"); requested != "" && i18n.Normalize(requested) == "" {
			respondError(c, content.Fail("UNSUPPORTED_LOCALE", 400))
			return
		}
		c.Next()
	}
}
func respondError(c *gin.Context, err error) {
	problem := content.ProblemOf(err)
	if problem.Status >= 500 {
		log.Printf("request_id=%s error=%v", c.Writer.Header().Get("X-Request-ID"), err)
	}
	c.AbortWithStatusJSON(problem.Status, gin.H{"error": gin.H{"code": problem.Code, "message": i18n.Message(locale(c), problem.Code)}, "request_id": c.Writer.Header().Get("X-Request-ID")})
}
func cors(origins []string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, origin := range origins {
		if value := strings.TrimSpace(origin); value != "" {
			allowed[value] = true
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if !allowed[origin] {
				respondError(c, content.Fail("ORIGIN_NOT_ALLOWED", 403))
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
func serveAsset(c *gin.Context, asset content.Asset) {
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	if asset.MIME == "application/octet-stream" {
		c.Header("Content-Disposition", "attachment")
	}
	c.Data(200, asset.MIME, asset.Data)
}
