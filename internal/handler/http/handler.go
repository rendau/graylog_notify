package http

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/rendau/graylog_notify/internal/service/core"
)

type Handler struct {
	core *core.Service
}

func New(core *core.Service) *Handler {
	return &Handler{
		core: core,
	}
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("fail to read body", slog.String("error", err.Error()))
		return
	}

	h.core.Send(body)
}
