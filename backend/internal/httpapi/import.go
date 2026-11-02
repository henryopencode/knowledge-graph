package httpapi

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"knowledgegraph/backend/internal/content"
	"knowledgegraph/backend/internal/i18n"
)

func (w *workspace) importDocuments(c *gin.Context) {
	var payload struct {
		Files []struct{ Name, Content string }
	}
	if err := c.ShouldBindJSON(&payload); err != nil || len(payload.Files) == 0 || len(payload.Files) > 50 {
		w.error(c, content.Fail("INVALID_REQUEST", 400))
		return
	}
	imported := []string{}
	failures := []gin.H{}
	for _, file := range payload.Files {
		if len(file.Content) > 512_000 {
			failures = append(failures, gin.H{"name": file.Name, "error": i18n.Message(locale(c), "DOCUMENT_TOO_LARGE"), "code": "DOCUMENT_TOO_LARGE"})
			continue
		}
		meta, body, err := content.LegacyMarkdown([]byte(file.Content))
		if err != nil {
			failures = append(failures, gin.H{"name": file.Name, "error": i18n.Message(locale(c), "INVALID_DOCUMENT"), "code": "INVALID_DOCUMENT"})
			continue
		}
		title, _ := meta["title"].(string)
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(file.Name), filepath.Ext(file.Name))
		}
		kind, _ := meta["kind"].(string)
		if !content.Kinds[kind] {
			kind = "concept"
		}
		tags := []string{}
		if raw, ok := meta["tags"].([]any); ok {
			for _, item := range raw {
				if tag, ok := item.(string); ok {
					tags = append(tags, tag)
				}
			}
		}
		// Import never grants publication, even if incoming frontmatter requests it.
		d, err := w.config.Store.Write(c.Request.Context(), "", content.WriteInput{Title: title, Kind: kind, Tags: tags, Content: body, Locale: locale(c), Channels: &content.Channels{Graph: true}})
		if err != nil {
			problem := content.ProblemOf(err)
			failures = append(failures, gin.H{"name": file.Name, "error": i18n.Message(locale(c), problem.Code), "code": problem.Code})
		} else {
			imported = append(imported, d.ID)
		}
	}
	c.JSON(200, gin.H{"imported": imported, "errors": failures})
}
