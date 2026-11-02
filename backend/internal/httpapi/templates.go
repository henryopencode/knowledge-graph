package httpapi

import "knowledgegraph/backend/internal/i18n"

func templates(locale string) map[string]string {
	result := map[string]string{}
	for _, kind := range []string{"project", "incident", "decision", "concept"} {
		result[kind] = i18n.Message(locale, "template."+kind)
	}
	return result
}
