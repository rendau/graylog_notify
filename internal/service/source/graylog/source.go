package graylog

import (
	"fmt"

	"github.com/goccy/go-json"

	"github.com/rendau/graylog_notify/internal/service/core"
)

type Source struct {
}

func New() *Source {
	return &Source{}
}

func (s *Source) ParseMessage(data []byte) (*core.Message, error) {
	message := &Message{}

	// fmt.Println("GraylogData:\n" + string(data))

	// parse root object
	err := json.Unmarshal(data, message)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	messageMap := map[string]any{}

	// parse fields.message object
	err = json.Unmarshal(message.Event.Fields.Message, &messageMap)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal Fields.Message: %w", err)
	}

	return core.NewMessage(message.Event.Fields.Tag, messageMap), nil
}
