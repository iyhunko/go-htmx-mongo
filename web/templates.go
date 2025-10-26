package web

import (
	"html/template"
)

// LoadTemplates loads all HTML templates with custom functions
func LoadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key := values[i].(string)
				dict[key] = values[i+1]
			}
			return dict
		},
	}

	return template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html")
}
