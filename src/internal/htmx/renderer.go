package htmx

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
)

type Page struct {
	Title string
	Data  any
}

type Renderer struct {
	base *template.Template
	sets map[string]*template.Template
}

func NewRenderer(uiFS embed.FS) *Renderer {

	t := template.Must(template.ParseFS(uiFS, "templates/**/*.html"))
	return &Renderer{
		base: t,
		sets: make(map[string]*template.Template),
	}
}

func (r *Renderer) Register(key string, fs embed.FS, patterns ...string) error {
	if _, exists := r.sets[key]; exists {
		return fmt.Errorf("renderer: page %q already registered", key)
	}

	t, err := r.base.Clone()
	if err != nil {
		return fmt.Errorf("renderr failed to clone base with error: %w", err)
	}

	t, err = t.ParseFS(fs, patterns...)
	if err != nil {
		return fmt.Errorf("renderer failed to parse fs with error: %w", err)
	}
	r.sets[key] = t
	return nil
}

func (r *Renderer) Render(ctx context.Context, key string, p Page,
	w http.ResponseWriter, req *http.Request,

) {

	var buf bytes.Buffer

	t, exists := r.sets[key]
	if !exists {
		http.Error(w, "set does not exist", 500)
		return
	}

	if err := t.ExecuteTemplate(&buf, "layout", p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	buf.WriteTo(w)
}

func (r *Renderer) Error(ctx context.Context, status int, msg string, w http.ResponseWriter, req *http.Request) {

	p := Page{
		Title: "Error",
		Data: map[string]any{"Error": msg,
			"StatusCode": status,
		},
	}
	var buf bytes.Buffer
	if err := r.base.ExecuteTemplate(&buf, "error", p); err != nil {
		http.Error(w, err.Error(), 500)
		return

	}
	w.WriteHeader(status)
	buf.WriteTo(w)

}
