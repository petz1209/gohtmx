package home

import (
	"bikes/internal/ui"
	"database/sql"
	"net/http"
)

// der router bindet das feature in unsere webapp ein.
// baut das gesamte feature zusammen und verknüpft dann die einzelnen requests
// mit spezifischen handlern.

// Register wires the home feature into the app.
func Register(mux *http.ServeMux, app *ui.App, db *sql.DB) error {

	view, err := app.NewFeature("home", "Home", files)
	if err != nil {
		return err
	}

	// hier registrieren wir defakto den renderer
	homePage, err := view.Page("index")
	if err != nil {
		return err
	}

	h := newHandler(homePage, newService())

	// hier registrieren wir die feature custom static files (also alles was im static ordner liegt)
	// in index.html werden diese dann im head referenziert via asset.
	// das ist optional, falls das feature eigene js und css hat die dazu geladen werden sollen.
	mux.Handle("GET "+view.StaticPrefix(), view.StaticHandler())
	mux.HandleFunc("GET /{$}", h.HomePage)
	return nil
}
