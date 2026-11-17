package main

import (
	"bikes/internal/features/riders"
	"net/http"
)

type handlers struct {
	Rider *riders.RiderHandler
}

func wireRoutes(mux *http.ServeMux, h *handlers) {

	mux.HandleFunc("GET /{$}", h.Rider.Index)
	mux.HandleFunc("GET /riders", h.Rider.Index)
	mux.HandleFunc("GET /riders/{id}", h.Rider.PagesDetail)
	mux.HandleFunc("GET /error", h.Rider.Error)

}
