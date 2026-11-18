package home

import (
	"bikes/internal/ui"
	"fmt"
	"net/http"
)

type handler struct {
	renderer *ui.Renderer
	srv      *service
}

func NewHandler(renderer *ui.Renderer, srv *service) *handler {

	return &handler{renderer: renderer, srv: srv}
}

func (h *handler) HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello")
	h.renderer.Render(r.Context(), ui.Page{}, w, r)

}
