package riders

import (
	"bikes/internal/ui"
	"database/sql"
	"net/http"
)

func Register(mux *http.ServeMux, app *ui.App, db *sql.DB) error {

	view, err := app.NewFeature("riders", "Riders", files)
	if err != nil {
		return err
	}

	overviewPage, err := view.Page("overview")
	if err != nil {
		return err
	}

	detailPage, err := view.Page("detail")
	if err != nil {
		return err
	}

	h := newHandler(overviewPage, detailPage, newService(newRepo(db)))

	mux.HandleFunc("GET /riders", h.OverviewPage)
	mux.HandleFunc("GET /riders/{id}", h.DetailPage)
	return nil

}
