package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"time"
)

//go:embed *.html static/*
var templateFS embed.FS

type TemplateEngine struct {
	templates map[string]*template.Template
}

// StaticFS returns an http.FileSystem serving embedded static assets (/static/).
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(templateFS, "static")
	if err != nil {
		return http.FS(templateFS)
	}
	return http.FS(sub)
}

func New(timezoneStr string) (*TemplateEngine, error) {
	if timezoneStr == "" {
		timezoneStr = "Local"
	}
	loc, err := time.LoadLocation(timezoneStr)
	if err != nil {
		loc = time.Local
	}

	engine := &TemplateEngine{
		templates: make(map[string]*template.Template),
	}

	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.In(loc).Format("2006-01-02 15:04:05 MST")
		},
		"formatDuration": func(start, finish any) string {
			var startTime, finishTime time.Time
			switch v := start.(type) {
			case time.Time:
				startTime = v
			}
			switch v := finish.(type) {
			case time.Time:
				finishTime = v
			}

			if startTime.IsZero() {
				return "-"
			}
			if finishTime.IsZero() {
				finishTime = time.Now()
			}
			d := finishTime.Sub(startTime)
			if d < 0 {
				d = 0
			}
			m := int(d.Minutes())
			s := int(d.Seconds()) % 60
			return fmt.Sprintf("%dm %ds", m, s)
		},
	}

	// Standalone pages (full HTML pages that do not extend layout.html)
	standalonePages := []string{
		"login.html",
	}

	for _, page := range standalonePages {
		tmpl, err := template.New(page).Funcs(funcMap).ParseFS(templateFS, page)
		if err != nil {
			return nil, fmt.Errorf("parse standalone template %s: %w", page, err)
		}
		engine.templates[page] = tmpl
	}

	// Pages extending layout.html
	layoutPages := []string{
		"dashboard.html",
		"tenants.html",
		"hosts.html",
		"apps.html",
		"runs.html",
		"console.html",
		"search.html",
		"users.html",
		"api_keys.html",
		"settings.html",
	}

	for _, page := range layoutPages {
		tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFS(templateFS, "layout.html", page)
		if err != nil {
			return nil, fmt.Errorf("parse layout template %s: %w", page, err)
		}
		engine.templates[page] = tmpl
	}

	return engine, nil
}

func (t *TemplateEngine) Render(w io.Writer, page string, data any) error {
	tmpl, ok := t.templates[page]
	if !ok {
		return fmt.Errorf("template %s not found", page)
	}
	return tmpl.Execute(w, data)
}

func (t *TemplateEngine) RenderPartial(w io.Writer, page, blockName string, data any) error {
	tmpl, ok := t.templates[page]
	if !ok {
		return fmt.Errorf("template %s not found", page)
	}
	return tmpl.ExecuteTemplate(w, blockName, data)
}
