package core

import (
	"encoding/json"
	"log/slog"

	"github.com/rendau/graylog_notify/internal/cns"
)

type SendRequestSt struct {
	Tag     string
	Message string
}

func (c *Core) Send(obj SendRequestSt) {
	msgMap := map[string]any{}

	if obj.Message != "" {
		err := json.Unmarshal([]byte(obj.Message), &msgMap)
		if err != nil {
			msgMap[cns.MessageFieldName] = obj.Message
		}
	}

	if obj.Tag != "" {
		msgMap[cns.TagFieldName] = obj.Tag
	}

	//slog.Info("Message", slog.Any("msgMap", msgMap))

	if c.destination != nil {
		err := c.destination.Send(msgMap)
		if err != nil {
			slog.Error("fail to destination.send", slog.String("error", err.Error()))
		}
	}
}
