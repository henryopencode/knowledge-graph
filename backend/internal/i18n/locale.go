package i18n

import (
	"embed"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

//go:embed messages/*.json
var files embed.FS
var catalogs = load()
var Supported = []string{"zh-CN", "zh-TW", "en"}

const Default = "zh-CN"

func load() map[string]map[string]string {
	result := map[string]map[string]string{}
	for _, locale := range []string{"zh-CN", "zh-TW", "en"} {
		raw, err := files.ReadFile("messages/" + locale + ".json")
		if err != nil {
			panic(err)
		}
		var messages map[string]string
		if err = json.Unmarshal(raw, &messages); err != nil {
			panic(err)
		}
		result[locale] = messages
	}
	for locale, messages := range result {
		for key := range result[Default] {
			if messages[key] == "" {
				panic("missing translation: " + locale + "/" + key)
			}
		}
		if len(messages) != len(result[Default]) {
			panic("unexpected translation keys: " + locale)
		}
	}
	return result
}
func Normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch {
	case value == "zh-tw" || value == "zh-hk" || strings.HasPrefix(value, "zh-hant"):
		return "zh-TW"
	case value == "zh" || value == "zh-cn" || value == "zh-sg" || strings.HasPrefix(value, "zh-hans"):
		return "zh-CN"
	case value == "en" || strings.HasPrefix(value, "en-"):
		return "en"
	}
	return ""
}
func Valid(value string) bool { return value == "zh-CN" || value == "zh-TW" || value == "en" }
func Negotiate(header string) string {
	type preference struct {
		locale  string
		quality float64
	}
	preferences := []preference{}
	for _, part := range strings.Split(header, ",") {
		bits := strings.Split(strings.TrimSpace(part), ";")
		quality := 1.0
		for _, bit := range bits[1:] {
			if strings.HasPrefix(strings.TrimSpace(bit), "q=") {
				if q, err := strconv.ParseFloat(strings.TrimPrefix(strings.TrimSpace(bit), "q="), 64); err == nil {
					quality = q
				}
			}
		}
		locale := Normalize(bits[0])
		if locale != "" && quality > 0 {
			preferences = append(preferences, preference{locale, quality})
		}
	}
	sort.SliceStable(preferences, func(i, j int) bool { return preferences[i].quality > preferences[j].quality })
	if len(preferences) > 0 {
		return preferences[0].locale
	}
	return Default
}
func FromRequest(r *http.Request) string {
	for _, key := range []string{"locale", "lang"} {
		if locale := Normalize(r.URL.Query().Get(key)); locale != "" {
			return locale
		}
	}
	if cookie, err := r.Cookie("henryweb_locale"); err == nil {
		if locale := Normalize(cookie.Value); locale != "" {
			return locale
		}
	}
	return Negotiate(r.Header.Get("Accept-Language"))
}
func Message(locale, key string) string {
	if !Valid(locale) {
		locale = Default
	}
	if text := catalogs[locale][key]; text != "" {
		return text
	}
	if text := catalogs[Default][key]; text != "" {
		return text
	}
	return key
}
func Catalog(locale string) map[string]string {
	if !Valid(locale) {
		locale = Default
	}
	copy := map[string]string{}
	for k, v := range catalogs[locale] {
		copy[k] = v
	}
	return copy
}
