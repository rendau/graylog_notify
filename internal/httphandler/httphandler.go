package httphandler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rendau/graylog_notify/internal/core"
)

type handlerSt struct {
	cr *core.Core
}

func NewHttpHandler(cr *core.Core) http.Handler {
	r := chi.NewRouter()

	// healthcheck
	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// --------------------

	h := &handlerSt{
		cr: cr,
	}

	// save
	r.Post("/send", h.Send)

	return r
}
