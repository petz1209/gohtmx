package ui

import (
	"io/fs"
	"net/http"
	"strings"
)

/*
this is used to register features spezific javascript files.
it is kind of nice because we can keep the spezifics at the feature level.
*/

// here we register custom js scripts
// this is nice because it allows us to simply add custom javascript
type StaticRegistry struct {
	baseStatic fs.FS
	jsFiles    map[string]fs.FS
}

// NewRegistry erwartet das FS ab dem "static"-Ordner (via fs.Sub)
func NewRegistry(baseStatic fs.FS) *StaticRegistry {
	return &StaticRegistry{
		baseStatic: baseStatic,
		jsFiles:    make(map[string]fs.FS),
	}
}

// RegisterJS registriert eine JS-Datei unter dem Zielnamen (z. B. "home.js")
func (r *StaticRegistry) RegisterJS(publicName string, sourceFS fs.FS) {
	r.jsFiles[publicName] = sourceFS
}

// Handler liefert sowohl globale Static-Files als auch registrierte Feature-Scripts aus
func (r *StaticRegistry) Handler() http.Handler {
	baseServer := http.FileServer(http.FS(r.baseStatic))

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Prüfen, ob ein JS-File angefragt wird (Pfad nach http.StripPrefix: "js/home.js")
		if strings.HasPrefix(req.URL.Path, "js/") {
			fileName := strings.TrimPrefix(req.URL.Path, "js/")

			// Wenn das File in der Registry liegt, direkt von dort ausliefern
			if featureFS, exists := r.jsFiles[fileName]; exists {
				// Pfad für den FileServer auf den Dateinamen anpassen
				req.URL.Path = "/" + fileName
				http.FileServer(http.FS(featureFS)).ServeHTTP(w, req)
				return
			}
		}

		// Fallback: Alles andere (inkl. static/js/main.js) kommt aus dem normalen static-Ordner
		baseServer.ServeHTTP(w, req)
	})
}
