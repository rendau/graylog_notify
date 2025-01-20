package http

import (
	"net/http"
)

func AssignRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /send", h.Send)
}
