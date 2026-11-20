package ui

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"runtime/debug"
)

// StaticPrefix is where StaticHandler serves the global static files.
const StaticPrefix = "/static/"

// StaticHandler serves the global static files (app.css, htmx, ...).
// The caller mounts it, so ui stays independent of the router:
//
//	mux.Handle("GET "+ui.StaticPrefix, ui.StaticHandler())
func StaticHandler() http.Handler {
	static, err := fs.Sub(files, "static")
	if err != nil {
		panic(err) // only fails if the embed directive is wrong
	}
	return http.StripPrefix(StaticPrefix, http.FileServerFS(static))
}

// App holds the global templates. Features are built from it (see NewFeature).
// It knows nothing about routing.
type App struct {
	base     *template.Template // global templates, cloned for every page
	errorSet *template.Template // eigene Kopie, nur zum Rendern der Fehlerseite
}

// NewApp parses the global templates.
func NewApp() (*App, error) {
	// "asset" must exist at parse time; each feature replaces it with its own version.
	base := template.New("ui").
		Funcs(funcMap).
		Funcs(template.FuncMap{"asset": assetFunc("")})
	base, err := parseFiles(base, files, findHTML(files, "templates", nil))
	if err != nil {
		return nil, fmt.Errorf("ui: global templates: %w", err)
	}
	errorSet, err := base.Clone()
	if err != nil {
		return nil, err
	}
	return &App{base: base, errorSet: errorSet}, nil
}

// RenderError rendert die globale Fehlerseite mit dem richtigen Statuscode.
func (a *App) RenderError(w http.ResponseWriter, r *http.Request, status int) {
	var buf bytes.Buffer
	data := PageError{Title: http.StatusText(status), StatusCode: status, Error: http.StatusText(status)}
	if err := a.errorSet.ExecuteTemplate(&buf, "error", data); err != nil {
		slog.ErrorContext(r.Context(), "ui: render error page", "err", err)
		http.Error(w, http.StatusText(status), status) // letzter Ausweg: Plain Text
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// findHTML returns every .html file below root (any depth) for which keep
// returns true (nil keeps all).
func findHTML(fsys fs.FS, root string, keep func(p string) bool) []string {
	var paths []string
	fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && path.Ext(p) == ".html" && (keep == nil || keep(p)) {
			paths = append(paths, p)
		}
		return nil
	})
	return paths
}

// parseFiles adds the given files to t; no files is not an error.
func parseFiles(t *template.Template, fsys fs.FS, paths []string) (*template.Template, error) {
	if len(paths) == 0 {
		return t, nil
	}
	return t.ParseFS(fsys, paths...)
}

func assetFunc(prefix string) func(string) string {
	return func(file string) string { return path.Join(prefix, file) }
}

func (a *App) Recover(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &trackingWriter{ResponseWriter: w}
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)

			}
			slog.ErrorContext(r.Context(), "panic recovered",
				"panic", rec, "path", r.URL.Path, "stack", string(debug.Stack()))
			if tw.wrote {
				return
			}
			a.RenderError(w, r, http.StatusInternalServerError)
		}()
		next.ServeHTTP(tw, r)

	})

}

// trackingWriter merkt sich, ob schon etwas geschrieben wurde.
type trackingWriter struct {
	http.ResponseWriter
	wrote bool
}

func (t *trackingWriter) WriteHeader(code int) { t.wrote = true; t.ResponseWriter.WriteHeader(code) }
func (t *trackingWriter) Write(b []byte) (int, error) {
	t.wrote = true
	return t.ResponseWriter.Write(b)
}
