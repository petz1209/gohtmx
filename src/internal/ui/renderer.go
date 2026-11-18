package ui

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net/http"
	"text/template"
)

type Renderer struct {
	assets *template.Template
}

// NewRenderer
// creates a new renderer for a given page or usecase.
// I think that should be fine to be honest.
func NewRenderer(globTemplates embed.FS, locTemplates embed.FS, patterns ...string) (*Renderer, error) {
	if patterns == nil {
		patterns = []string{"templates/**/*.html"}
	}

	// here lets register the gobal and local templates.
	// this makes live easier (i think)

	t := template.Must(template.ParseFS(globTemplates, "templates/**/*.html"))
	t, err := t.ParseFS(locTemplates, patterns...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse localTeamplates with err: %w", err)
	}

	return &Renderer{assets: t}, nil
}

func (h *Renderer) Render(ctx context.Context, p Page,
	w http.ResponseWriter, req *http.Request,

) {

	var buf bytes.Buffer

	if err := h.assets.ExecuteTemplate(&buf, "layout", p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(200)
	buf.WriteTo(w)
}
