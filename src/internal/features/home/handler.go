package home

import (
	"bikes/internal/ui"
	"log/slog"
	"net/http"
)

// der handler ist in diesem fall der ui handler.
// er nimmt den request entgegen erarbeitet ein ergebnis
// und gibt das ergebnis weiter an den jeweiligen ui.Page teil
// oder wenn es nur ein partial ist an irrgendwas anderes.

type handler struct {
	homePage *ui.Page
	srv      *service
}

func newHandler(homePage *ui.Page, srv *service) *handler {
	return &handler{homePage: homePage, srv: srv}
}

func (h *handler) HomePage(w http.ResponseWriter, r *http.Request) {
	// with restify this whole block is just: return h.homePage.Render(w, r, ui.PageData{})
	if err := h.homePage.Render(w, r, ui.PageData{}); err != nil {
		slog.ErrorContext(r.Context(), "home page", "err", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
