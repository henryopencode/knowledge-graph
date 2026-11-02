package httpapi

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"knowledgegraph/backend/internal/auth"
	"knowledgegraph/backend/internal/content"
	"knowledgegraph/backend/internal/i18n"
)

type WorkspaceConfig struct {
	Store     *content.Store
	Auth      *auth.Auth
	AssetsDir string
}
type workspace struct {
	config WorkspaceConfig
	pages  map[string]*template.Template
}

func NewWorkspace(config WorkspaceConfig) (*gin.Engine, error) {
	app := &workspace{config: config, pages: map[string]*template.Template{}}
	for _, language := range i18n.Supported {
		for _, name := range []string{"index.html", "login.html"} {
			raw, err := os.ReadFile(filepath.Join(config.AssetsDir, name))
			if err != nil {
				return nil, err
			}
			selected := language
			page, err := template.New(name).Funcs(template.FuncMap{"t": func(key string) string { return i18n.Message(selected, "ui."+key) }}).Parse(string(raw))
			if err != nil {
				return nil, err
			}
			app.pages[language+"/"+name] = page
		}
	}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), language(), func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2_100_000)
		c.Next()
	})
	if err := router.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		return nil, err
	}
	router.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "henryweb-workspace"}) })
	router.GET("/", app.index)
	router.GET("/index.html", app.index)
	router.GET("/i18n.js", app.i18nScript)
	for _, name := range []string{"login.css", "login.js", "favicon.svg"} {
		asset := name
		router.GET("/"+asset, func(c *gin.Context) { c.File(filepath.Join(config.AssetsDir, asset)) })
	}
	router.POST("/api/login", func(c *gin.Context) {
		if !config.Auth.AllowedMutation(c.Request) {
			app.error(c, content.Fail("ORIGIN_NOT_ALLOWED", 403))
			return
		}
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			app.error(c, content.Fail("INVALID_REQUEST", 400))
			return
		}
		token, err := config.Auth.Login(input.Username, input.Password, c.ClientIP())
		if err != nil {
			var failure *auth.Error
			if errors.As(err, &failure) {
				app.error(c, content.Fail(failure.Code, failure.Status))
			} else {
				app.error(c, err)
			}
			return
		}
		http.SetCookie(c.Writer, config.Auth.Cookie(c.Request, token, false))
		c.JSON(200, gin.H{"ok": true})
	})
	protected := router.Group("/")
	protected.Use(func(c *gin.Context) {
		if !config.Auth.Authenticated(c.Request) {
			app.error(c, content.Fail("AUTH_REQUIRED", 401))
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && !config.Auth.AllowedMutation(c.Request) {
			app.error(c, content.Fail("ORIGIN_NOT_ALLOWED", 403))
			return
		}
		c.Next()
	})
	for _, name := range []string{"app.js", "navigation.js", "styles.css"} {
		asset := name
		protected.GET("/"+asset, func(c *gin.Context) { c.File(filepath.Join(config.AssetsDir, asset)) })
	}
	protected.POST("/api/logout", func(c *gin.Context) {
		config.Auth.Logout(c.Request)
		http.SetCookie(c.Writer, config.Auth.Cookie(c.Request, "", true))
		c.JSON(200, gin.H{"ok": true})
	})
	protected.GET("/api/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	protected.GET("/api/notes", app.list)
	protected.POST("/api/notes", app.create)
	protected.PUT("/api/notes/:id", app.update)
	protected.DELETE("/api/notes/:id", app.remove)
	protected.GET("/api/export", app.export)
	protected.POST("/api/import", app.importDocuments)
	private := protected.Group("/api/v1/private")
	private.GET("/documents", app.list)
	private.GET("/documents/:id", app.get)
	private.POST("/documents", app.create)
	private.PUT("/documents/:id", app.update)
	private.DELETE("/documents/:id", app.remove)
	private.POST("/documents/:id/publish", app.publish)
	private.POST("/documents/:id/withdraw", app.withdraw)
	private.POST("/documents/:id/preview", app.preview)
	private.GET("/documents/:id/assets/*path", func(c *gin.Context) {
		asset, err := config.Store.PrivateAsset(c.Param("id"), strings.TrimPrefix(c.Param("path"), "/"))
		if err != nil {
			app.error(c, err)
			return
		}
		serveAsset(c, asset)
	})
	router.NoRoute(func(c *gin.Context) { app.error(c, content.Fail("NOT_FOUND", 404)) })
	return router, nil
}
func (w *workspace) error(c *gin.Context, err error) {
	if strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
		respondError(c, err)
		return
	}
	problem := content.ProblemOf(err)
	c.AbortWithStatusJSON(problem.Status, gin.H{"error": i18n.Message(locale(c), problem.Code), "code": problem.Code})
}
func (w *workspace) index(c *gin.Context) {
	name := "login.html"
	if w.config.Auth.Authenticated(c.Request) {
		name = "index.html"
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: "henryweb_locale", Value: locale(c), Path: "/", MaxAge: 31536000, SameSite: http.SameSiteLaxMode})
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := w.pages[locale(c)+"/"+name].Execute(c.Writer, struct{ Locale string }{locale(c)}); err != nil {
		_ = c.Error(err)
	}
}
func (w *workspace) i18nScript(c *gin.Context) {
	payload, _ := json.Marshal(gin.H{"locale": locale(c), "messages": i18n.Catalog(locale(c))})
	source := `"use strict";globalThis.WorkspaceI18n=(()=>{const data=` + string(payload) + `;const t=(key,values={})=>{const count=values.count??values.p0;const plural=typeof count==='number'&&new Intl.PluralRules(data.locale).select(count)==='one';const text=(plural&&data.messages['ui.'+key+'_one'])||data.messages['ui.'+key]||key;return text.replace(/\{(\w+)\}/g,(_,name)=>String(values[name]??''));};return {locale:data.locale,t};})();`
	c.Data(200, "application/javascript; charset=utf-8", []byte(source))
}
func (w *workspace) documents(c *gin.Context) ([]content.Document, error) {
	return w.config.Store.List(locale(c), false)
}
func (w *workspace) list(c *gin.Context) {
	documents, err := w.documents(c)
	if err != nil {
		w.error(c, err)
		return
	}
	notes := []gin.H{}
	for _, d := range documents {
		encoded, _ := json.Marshal(d)
		var note gin.H
		_ = json.Unmarshal(encoded, &note)
		html, err := w.config.Store.RenderPrivate(d, documents)
		if err != nil {
			w.error(c, err)
			return
		}
		note["html"] = html
		notes = append(notes, note)
	}
	c.JSON(200, gin.H{"notes": notes, "templates": templates(locale(c)), "vault": "documents", "storage": "server", "authenticated": true, "locale": locale(c), "locales": i18n.Supported})
}
func (w *workspace) get(c *gin.Context) {
	d, err := w.config.Store.Get(c.Param("id"), locale(c))
	if err != nil {
		w.error(c, err)
		return
	}
	documents, err := w.documents(c)
	if err != nil {
		w.error(c, err)
		return
	}
	for _, candidate := range documents {
		if candidate.ID == d.ID && candidate.Locale == d.Locale {
			d = candidate
		}
	}
	c.JSON(200, d)
}
func (w *workspace) input(c *gin.Context) (content.WriteInput, bool) {
	var input content.WriteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		w.error(c, content.Fail("INVALID_REQUEST", 400))
		return input, false
	}
	if input.Locale == "" {
		input.Locale = locale(c)
	}
	return input, true
}
func (w *workspace) create(c *gin.Context) {
	input, ok := w.input(c)
	if !ok {
		return
	}
	d, err := w.config.Store.Write(c.Request.Context(), "", input)
	if err != nil {
		w.error(c, err)
		return
	}
	c.JSON(201, d)
}
func (w *workspace) update(c *gin.Context) {
	input, ok := w.input(c)
	if !ok {
		return
	}
	d, err := w.config.Store.Write(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		w.error(c, err)
		return
	}
	c.JSON(200, d)
}
func (w *workspace) remove(c *gin.Context) {
	var input struct{ Version, Locale string }
	if err := c.ShouldBindJSON(&input); err != nil {
		w.error(c, content.Fail("INVALID_REQUEST", 400))
		return
	}
	if input.Locale == "" {
		input.Locale = locale(c)
	}
	if err := w.config.Store.Remove(c.Request.Context(), c.Param("id"), input.Locale, input.Version); err != nil {
		w.error(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
func (w *workspace) publicationInput(c *gin.Context) (string, string, bool) {
	var input struct {
		Locale  string `json:"locale"`
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		w.error(c, content.Fail("INVALID_REQUEST", 400))
		return "", "", false
	}
	if input.Locale == "" {
		input.Locale = locale(c)
	}
	return input.Locale, input.Version, true
}
func (w *workspace) publish(c *gin.Context) {
	language, version, ok := w.publicationInput(c)
	if !ok {
		return
	}
	post, err := w.config.Store.Publish(c.Request.Context(), c.Param("id"), language, version)
	if err != nil {
		w.error(c, err)
		return
	}
	c.JSON(200, post)
}
func (w *workspace) withdraw(c *gin.Context) {
	language, version, ok := w.publicationInput(c)
	if !ok {
		return
	}
	if err := w.config.Store.Withdraw(c.Request.Context(), c.Param("id"), language, version); err != nil {
		w.error(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}
func (w *workspace) preview(c *gin.Context) {
	language, version, ok := w.publicationInput(c)
	if !ok {
		return
	}
	post, err := w.config.Store.Preview(c.Request.Context(), c.Param("id"), language, version)
	if err != nil {
		w.error(c, err)
		return
	}
	c.JSON(200, post)
}
func (w *workspace) export(c *gin.Context) {
	data, err := w.config.Store.Export()
	if err != nil {
		w.error(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="henryweb-documents.zip"`)
	c.Data(200, "application/zip", data)
}
