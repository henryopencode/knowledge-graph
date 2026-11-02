package content

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Channels struct {
	Blog  bool `json:"blog" yaml:"blog"`
	Graph bool `json:"graph" yaml:"graph"`
}

// Metadata is the versioned, portable contract written in each Markdown file.
type Metadata struct {
	SchemaVersion  int            `json:"schema_version" yaml:"schema_version"`
	ID             string         `json:"id" yaml:"id"`
	Locale         string         `json:"locale" yaml:"locale"`
	Title          string         `json:"title" yaml:"title"`
	Summary        string         `json:"summary" yaml:"summary"`
	Slug           string         `json:"slug" yaml:"slug"`
	Kind           string         `json:"kind" yaml:"kind"`
	Category       string         `json:"category" yaml:"category"`
	Tags           []string       `json:"tags" yaml:"tags"`
	Channels       Channels       `json:"channels" yaml:"channels"`
	Created        string         `json:"created" yaml:"created"`
	Updated        string         `json:"updated" yaml:"updated"`
	SourceLocale   string         `json:"source_locale,omitempty" yaml:"source_locale,omitempty"`
	SourceRevision string         `json:"source_revision,omitempty" yaml:"source_revision,omitempty"`
	LegacyID       string         `json:"legacy_id,omitempty" yaml:"legacy_id,omitempty"`
	Sample         bool           `json:"sample,omitempty" yaml:"sample,omitempty"`
	Properties     map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`
}

type Document struct {
	Metadata
	Content            string   `json:"content"`
	Version            string   `json:"version"`
	PublicationVersion string   `json:"publication_version"`
	Outgoing           []string `json:"outgoing"`
	Incoming           []string `json:"incoming"`
	Missing            []string `json:"missing"`
	Ambiguous          []string `json:"ambiguous"`
	AvailableLocales   []string `json:"available_locales"`
	TranslationStale   bool     `json:"translation_stale"`
	PublishedRevision  string   `json:"published_revision,omitempty"`
	PublishedAt        string   `json:"published_at,omitempty"`
}

type WriteInput struct {
	Title              string    `json:"title"`
	Summary            string    `json:"summary"`
	Slug               string    `json:"slug"`
	Kind               string    `json:"kind"`
	Category           string    `json:"category"`
	Tags               []string  `json:"tags"`
	Content            string    `json:"content"`
	Locale             string    `json:"locale"`
	Version            string    `json:"version"`
	PublicationVersion string    `json:"publication_version,omitempty"`
	Channels           *Channels `json:"channels,omitempty"`
	SourceLocale       string    `json:"source_locale,omitempty"`
	SourceRevision     string    `json:"source_revision,omitempty"`
}

// PublicPost deliberately excludes storage paths, working revisions and graph data.
type PublicPost struct {
	ID               string            `json:"id"`
	Locale           string            `json:"locale"`
	ContentLocale    string            `json:"content_locale"`
	AvailableLocales []string          `json:"available_locales"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Slug             string            `json:"slug"`
	Category         string            `json:"category"`
	Tags             []string          `json:"tags"`
	Date             string            `json:"date"`
	ReadingMinutes   int               `json:"reading_minutes"`
	HTML             string            `json:"html"`
	Revision         string            `json:"revision"`
	Translations     map[string]string `json:"translations"`
}

type Problem struct {
	Code   string
	Status int
}

func (p *Problem) Error() string         { return p.Code }
func Fail(code string, status int) error { return &Problem{Code: code, Status: status} }
func ProblemOf(err error) *Problem {
	var problem *Problem
	if errors.As(err, &problem) {
		return problem
	}
	return &Problem{Code: "INTERNAL_ERROR", Status: http.StatusInternalServerError}
}

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,119}$`)
var Kinds = map[string]bool{"project": true, "incident": true, "decision": true, "concept": true}
var Locales = []string{"zh-CN", "zh-TW", "en"}

func ValidLocale(locale string) bool {
	for _, supported := range Locales {
		if locale == supported {
			return true
		}
	}
	return false
}
func ValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil && id == strings.ToLower(id)
}
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func validate(d Document) error {
	if d.SchemaVersion != 1 || !ValidID(d.ID) {
		return Fail("INVALID_DOCUMENT", 400)
	}
	if !ValidLocale(d.Locale) {
		return Fail("UNSUPPORTED_LOCALE", 400)
	}
	if strings.TrimSpace(d.Title) == "" || len([]rune(d.Title)) > 150 || strings.ContainsAny(d.Title, "\r\n\x00") {
		return Fail("INVALID_TITLE", 400)
	}
	if !Kinds[d.Kind] || !slugPattern.MatchString(d.Slug) {
		return Fail("INVALID_DOCUMENT", 400)
	}
	if len(d.Content) > 512_000 || len(d.Tags) > 30 || len(d.Summary) > 2000 {
		return Fail("DOCUMENT_TOO_LARGE", 400)
	}
	for _, tag := range d.Tags {
		if len(tag) > 120 || strings.ContainsAny(tag, "\r\n\x00") {
			return Fail("INVALID_DOCUMENT", 400)
		}
	}
	if d.SourceLocale != "" && (!ValidLocale(d.SourceLocale) || d.SourceLocale == d.Locale || d.SourceRevision == "") {
		return Fail("INVALID_TRANSLATION", 400)
	}
	if _, err := time.Parse(time.RFC3339, d.Created); err != nil {
		return Fail("INVALID_DOCUMENT", 400)
	}
	if _, err := time.Parse(time.RFC3339, d.Updated); err != nil {
		return Fail("INVALID_DOCUMENT", 400)
	}
	return nil
}
