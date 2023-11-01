package tgram

import (
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rendau/graylog_notify/internal/cns"
)

const (
	maxMsgFieldValueSize = 500
)

type TGram struct {
	initChatId int64

	botApi *tgbotapi.BotAPI
}

func NewTGram(botToken string, initChatId int64) (*TGram, error) {
	var err error

	res := &TGram{
		initChatId: initChatId,
	}

	res.botApi, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		slog.Error("Fail to create telegram-bot", slog.String("error", err.Error()))
		return nil, err
	}

	return res, nil
}

func (o *TGram) Send(msg map[string]any) error {
	for k, v := range msg {
		switch val := v.(type) {
		case string:
			if len([]rune(val)) > maxMsgFieldValueSize {
				msg[k] = string(([]rune(val))[:maxMsgFieldValueSize]) + "..."
			}
		}
	}

	var tag string
	if v, ok := msg[cns.TagFieldName]; ok {
		tag = v.(string)
		delete(msg, cns.TagFieldName)
	}

	var level string
	if v, ok := msg["level"]; ok && v != nil {
		level = v.(string)
		delete(msg, "level")
	} else if v, ok = msg["lvl"]; ok && v != nil {
		level = v.(string)
		delete(msg, "lvl")
	}

	var ts string
	if v, ok := msg["time"]; ok && v != nil {
		ts = v.(string)
		delete(msg, "time")
	} else if v, ok = msg["ts"]; ok && v != nil {
		ts = v.(string)
		delete(msg, "ts")
	} else if v, ok = msg["timestamp"]; ok && v != nil {
		ts = v.(string)
		delete(msg, "timestamp")
	}
	if ts != "" {
		ts = strings.ReplaceAll(ts, "-", "\\-")
		ts = strings.ReplaceAll(ts, "+", "\\+")
	}

	msgContent := " `" + tag + "` \\(" + level + "\\):  " + ts + "\n\n```\n"

	var mes string
	if v, ok := msg["msg"]; ok && v != nil {
		mes = v.(string)
		delete(msg, "msg")
	} else if v, ok = msg["message"]; ok && v != nil {
		mes = v.(string)
		delete(msg, "message")
	}
	if mes != "" {
		msgContent += fmt.Sprintf("   MSG: \t\t%s\n", mes)
	}

	var er string
	if v, ok := msg["error"]; ok && v != nil {
		er = v.(string)
		delete(msg, "error")
	} else if v, ok = msg["err"]; ok && v != nil {
		er = v.(string)
		delete(msg, "err")
	}
	if er != "" {
		msgContent += fmt.Sprintf("   ERROR: \t\t%s\n", er)
	}

	var caller string
	if v, ok := msg["caller"]; ok && v != nil {
		caller = v.(string)
		delete(msg, "caller")
	}
	if caller != "" {
		msgContent += fmt.Sprintf("   CALLER: \t\t%s\n", caller)
	}

	for k, v := range msg {
		msgContent += fmt.Sprintf("   %s: \t\t%v\n", strings.ToUpper(k), v)
	}

	msgContent += "```"

	tgMsg := tgbotapi.NewMessage(o.initChatId, msgContent)
	tgMsg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := o.botApi.Send(tgMsg)

	return err
}
