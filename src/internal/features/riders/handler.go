package riders

import (
	"bikes/internal/features/riders/domain"
	"bikes/internal/features/riders/repository"
	"bikes/internal/htmx"
	"database/sql"
	"embed"
	"log"
	"net/http"
)

//go:embed templates
var fs embed.FS

type RiderHandler struct {
	srv      *Service
	renderer *htmx.Renderer
}

func NewRiderHandler(
	renderer *htmx.Renderer,
	db *sql.DB) *RiderHandler {

	repo := repository.NewRiderRepository(db)
	service := NewService(repo)

	// register at the renderer
	if err := renderer.Register("rider", fs, "templates/*.html"); err != nil {
		log.Fatal(err)
	}
	if err := renderer.Register("rider.details", fs, "templates/details/*.html"); err != nil {
		log.Fatal(err)
	}

	return &RiderHandler{srv: service, renderer: renderer}
}

func (h *RiderHandler) Index(w http.ResponseWriter, r *http.Request) {

	data, err := h.srv.GetRiders(r.Context())
	if err != nil {
		http.Error(w, "internal server error", 500)
		return
	}

	p := htmx.Page{Title: "Riders", Data: data}
	h.renderer.Render(r.Context(), "rider", p, w, r)
	return
}

func (h *RiderHandler) PagesDetail(w http.ResponseWriter, r *http.Request) {
	const setname = "rider.details"

	p := htmx.Page{Title: "Riders", Data: domain.RiderDetails{}}
	h.renderer.Render(r.Context(), setname, p, w, r)
	return
}

func (h *RiderHandler) Error(w http.ResponseWriter, r *http.Request) {

	h.renderer.Error(r.Context(), 500, "An Error Occured", w, r)
	return
}
