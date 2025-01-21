package telegram

import (
	"fmt"
)

type Rule struct {
	Name   string
	ChatId int64
}

func (o *Rule) String() string {
	return fmt.Sprintf("{name: %s, chat_id: %d}", o.Name, o.ChatId)
}
