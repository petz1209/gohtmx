package home

import (
	"bikes/internal/ui"
	"embed"
	"io/fs"
	"net/http"
)

// instanciate a new router for the Home Feature
func NewHomeRouter(mux *http.ServeMux, globTemplate embed.FS, reg *ui.StaticRegistry) error {

	// lets setup the renderer.
	if err := RegisterAssets(reg); err != nil {
		return err
	}

	renderer, err := ui.NewRenderer("home", globTemplate, htmlFiles)
	if err != nil {
		return err
	}
	// after the renderer we setup the service
	srv := NewService()

	h := NewHandler(renderer, srv)
	// then we set up the handler.

	// finally we register the routes

	mux.HandleFunc("GET /{$}", h.HomePage)

	return nil

}

// register assets that you want in the system
func RegisterAssets(reg *ui.StaticRegistry) error {
	// Pfad auf den scripts-Ordner verkürzen, damit "home.js" im Root liegt
	sub, err := fs.Sub(jsFiles, "templates/scripts")
	if err != nil {
		return err
	}

	// Macht die Datei unter /static/js/home.js erreichbar
	reg.RegisterJS("home.js", sub)

	sub, err = fs.Sub(cssFiles, "templates/css")
	if err != nil {
		return err
	}
	reg.RegisterJS("home.css", sub)
	return nil
}
