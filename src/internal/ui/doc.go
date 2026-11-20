// Package ui is the central rendering framework. Features plug into it;
// they never deal with template parsing themselves.
//
// ui does no routing: it hands out http.Handlers and renders pages, the
// caller (std mux today, restify later) mounts and wraps them.
//
// # Setup (main.go)
//
//	app, err := ui.NewApp()
//	mux.Handle("GET "+ui.StaticPrefix, ui.StaticHandler())
//	home.Register(mux, app)
//
// # Adding a feature
//
//	features/<name>/
//	  assets.go        //go:embed templates static
//	                   var files embed.FS
//	  templates/
//	    pages/*.html   one file = one page, each must define "content"
//	    **/*.html      everything else: partials shared by all pages of the feature
//	  static/          optional, served under /assets/<name>/
//	  router.go        NewFeature, Page(...) for each page, mount static + routes
//	  handler.go       page.Render(w, r, ui.PageData{...})
//
// # Template sets
//
// Every page gets its own set:
//
//	global templates (ui/templates) + feature templates outside pages/ + the page file
//
// So every page can define the same block names without colliding.
// The global "layout" template calls:
//
//	"content"  required. The page body.
//	"head"     optional. Extra tags in <head> (CSS, JS, meta).
//
// Template data is always a ui.PageData; your own data goes in PageData.Data.
//
// # Template functions
//
//	{{asset "home.css"}}  -> /assets/<name>/home.css
//
// # Errors
//
// Everything that can be checked (syntax, missing "content", unknown page
// name) fails in NewFeature / Page, at startup. Page.Render returns an
// error and writes nothing in that case; the handler decides the response.
//
// Templates are parsed with html/template, so all output is escaped
// contextually.
package ui
