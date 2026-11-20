package ui

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

const pagesDir = "templates/pages/"

// Feature is a self-contained part of the UI: its pages, shared templates
// and static files. It knows nothing about routing; the feature's router
// mounts StaticHandler and the handlers itself.
type Feature struct {
	name   string
	prefix string // "/assets/<name>/"
	static http.Handler
	pages  map[string]*Page
}

// NewFeature parses all pages of the feature. Every page gets its own
// template set, so every page can define "content" and "head":
//
//	set = global templates + feature templates outside pages/ + one page file
//
// All errors (syntax, missing "content") surface here, at startup.
func (a *App) NewFeature(name, defaultTitle string, fsys fs.FS) (*Feature, error) {
	if name == "" || strings.ContainsAny(name, "/ ") {
		return nil, fmt.Errorf("ui: invalid feature name %q", name)
	}
	f := &Feature{
		name:   name,
		prefix: "/assets/" + name + "/",
		static: http.NotFoundHandler(),
		pages:  map[string]*Page{},
	}

	// feature set: global + everything except pages (partials, components, ...)
	shared, err := a.base.Clone()
	if err != nil {
		return nil, err
	}
	shared = shared.Funcs(template.FuncMap{"asset": assetFunc(f.prefix)})
	isPage := func(p string) bool { return strings.HasPrefix(p, pagesDir) }
	shared, err = parseFiles(shared, fsys, findHTML(fsys, "templates", func(p string) bool { return !isPage(p) }))
	if err != nil {
		return nil, fmt.Errorf("ui: feature %q templates: %w", name, err)
	}

	// one set per page: feature set + that page
	for _, file := range findHTML(fsys, "templates", isPage) {
		pageName := strings.TrimSuffix(strings.TrimPrefix(file, pagesDir), ".html")

		tmpl, err := shared.Clone()
		if err != nil {
			return nil, err
		}
		if tmpl, err = tmpl.ParseFS(fsys, file); err != nil {
			return nil, fmt.Errorf("ui: feature %q page %q: %w", name, pageName, err)
		}
		if tmpl.Lookup("content") == nil {
			return nil, fmt.Errorf("ui: feature %q page %q does not define \"content\"", name, pageName)
		}
		f.pages[pageName] = &Page{name: name + "/" + pageName, defaultTitle: defaultTitle, tmpl: tmpl}
	}

	if _, err := fs.Stat(fsys, "static"); err == nil {
		static, err := fs.Sub(fsys, "static")
		if err != nil {
			return nil, err
		}
		f.static = http.StripPrefix(f.prefix, http.FileServerFS(static))
	}

	return f, nil
}

// Page returns the page parsed from templates/pages/<name>.html.
// Call it at startup, so a typo fails there and not at request time.
func (f *Feature) Page(name string) (*Page, error) {
	p, ok := f.pages[name]
	if !ok {
		return nil, fmt.Errorf("ui: feature %q has no page %q (expected %s%s.html)", f.name, name, pagesDir, name)
	}
	return p, nil
}

// StaticPrefix is the URL prefix of the feature's static files: "/assets/<name>/".
func (f *Feature) StaticPrefix() string { return f.prefix }

// StaticHandler serves the feature's static/ dir. Mount it at StaticPrefix:
//
//	mux.Handle("GET "+f.StaticPrefix(), f.StaticHandler())
func (f *Feature) StaticHandler() http.Handler { return f.static }
