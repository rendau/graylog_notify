package telegram

import (
	"fmt"
)

type Rule struct {
	Tags    []string
	ChatIds []int64
}

func (r *Rule) String() string {
	return fmt.Sprintf("{tags: %v, chat_ids: %v}", r.Tags, r.ChatIds)
}
