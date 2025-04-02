package core

import (
	"log/slog"
	"time"
)

type Service struct {
	source      SourceI
	destination DestinationI
}

func New(
	source SourceI,
	destination DestinationI,
) *Service {
	return &Service{
		source:      source,
		destination: destination,
	}
}

func (s *Service) Send(data []byte) {
	msg, err := s.source.ParseMessage(data)
	if err != nil {
		slog.Error("fail to parse source message", slog.String("error", err.Error()), slog.Any("data", string(data)))
		err = s.destination.Send(&Message{
			Level:   "error",
			Ts:      time.Now().Format(time.RFC3339),
			Message: "GraylogNotify service: fail to parse source message",
			Error:   err.Error(),
			V:       make(map[string]any),
		})
		if err != nil {
			slog.Error("fail to destination.send", slog.String("error", err.Error()))
		}
		return
	}

	err = s.destination.Send(msg)
	if err != nil {
		slog.Error("fail to destination.send", slog.String("error", err.Error()))
	}
}
