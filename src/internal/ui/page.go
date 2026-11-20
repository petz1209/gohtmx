package ui

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// Page is one full page of a feature, with its own template set.
type Page struct {
	name         string // "<feature>/<page>", for error messages
	defaultTitle string
	tmpl         *template.Template
}

// Render renders p into the global layout.
//
// Rendering is buffered: on error nothing has been written, and the caller
// decides how to answer (with restify: just return the error).
func (pg *Page) Render(w http.ResponseWriter, r *http.Request, p PageData) error {
	if p.Title == "" {
		p.Title = pg.defaultTitle
	}

	var buf bytes.Buffer
	if err := pg.tmpl.ExecuteTemplate(&buf, "layout", p); err != nil {
		return fmt.Errorf("ui: render %s: %w", pg.name, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}

func (pg *Page) RenderErrorPage(w http.ResponseWriter, r *http.Request, p PageError) error {
	if p.Title == "" {
		p.Title = "error"
	}

	var buf bytes.Buffer
	if err := pg.tmpl.ExecuteTemplate(&buf, "error", p); err != nil {
		return fmt.Errorf("ui: render %s: %w", pg.name, err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}
