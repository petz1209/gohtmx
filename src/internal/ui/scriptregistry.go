package ui

import (
	"io/fs"
	"net/http"
	"path"
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

	// js and css files spezific to a custom page. zb /home   home.js,   home.css
	// this can run into naming conflicts so please name the files like your actual page.
	customFiles map[string]fs.FS
}

// NewRegistry erwartet das FS ab dem "static"-Ordner (via fs.Sub)
func NewRegistry(baseStatic fs.FS) *StaticRegistry {
	return &StaticRegistry{
		baseStatic:  baseStatic,
		customFiles: make(map[string]fs.FS),
	}
}

// RegisterJS registriert eine JS-Datei unter dem Zielnamen (z. B. "home.js")
func (r *StaticRegistry) RegisterJS(publicName string, sourceFS fs.FS) {
	r.customFiles[publicName] = sourceFS
}

func (r *StaticRegistry) Handler() http.Handler {
	//

	//baseServer := http.StripPrefix("/static/", http.FileServer(http.FS(r.baseStatic)))
	baseServer := http.FileServer(http.FS(r.baseStatic))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// "/static/css/home.css" -> "css/home.css"
		cleanPath := strings.TrimPrefix(req.URL.Path, "/static/")
		fileName := path.Base(req.URL.Path)
		contentType := getContentType(fileName)

		// hier checken wir ob das file in den custom files der features liegt.
		if featureFS, exists := r.customFiles[fileName]; exists {
			// Nur den reinen Dateinamen für fs.ReadFile extrahieren ("home.css")
			fileName := path.Base(cleanPath)

			content, err := fs.ReadFile(featureFS, fileName)
			if err != nil {
				http.NotFound(w, req)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(content)
			return
		}

		// Alle nicht registrierten Dateien kommen normal aus baseStatic
		baseServer.ServeHTTP(w, req)
		return

	})
}

// takes a fileName, checks the extension and returns the appropriate content type
func getContentType(fileName string) string {
	switch path.Ext(fileName) {
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}
