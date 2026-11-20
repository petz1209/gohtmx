package riders

import (
	"bikes/internal/ui"
	"fmt"
	"net/http"
	"strconv"
)

type handler struct {

	// renderer für die OverviewPage
	overviewpage *ui.Page
	// renderer für die DetailPage
	detailpage *ui.Page
	srv        *service
}

func newHandler(overviewpage, detailpage *ui.Page, srv *service) *handler {

	return &handler{overviewpage: overviewpage,
		detailpage: detailpage,
		srv:        srv}

}

func (h *handler) OverviewPage(w http.ResponseWriter, r *http.Request) {

	riders, err := h.srv.GetAllRiders(r.Context())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		e := h.overviewpage.RenderErrorPage(w, r, ui.PageError{
			Error: err.Error(), StatusCode: 500})
		if e != nil {
			fmt.Printf("Error Rendering: %v\n", e)
		}
		return
	}

	if riders == nil {
		e := h.overviewpage.RenderErrorPage(w, r, ui.PageError{
			Error: "No Riders found"})
		if e != nil {
			fmt.Printf("Error Rendering: %v\n", e)
		}
		return

	}
	e := h.overviewpage.Render(w, r, ui.PageData{Data: riders})
	if e != nil {
		fmt.Printf("Error Rendering: %v\n", e)
	}
	return
}

func (h *handler) DetailPage(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.detailpage.RenderErrorPage(w, r, ui.PageError{StatusCode: 400, Error: "invalid id"})
		return
	}

	rider, err := h.srv.repo.GetRiderDetails(r.Context(), id)
	if err != nil {
		h.detailpage.RenderErrorPage(w, r, ui.PageError{StatusCode: 404, Error: fmt.Sprintf("Rider not found for id: %d with error:%v", id, err)})
		return
	}

	h.detailpage.Render(w, r, ui.PageData{Data: rider, Title: fmt.Sprintf("%s %s", rider.FirstName, rider.LastName)})
	return
}
