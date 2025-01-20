package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rendau/graylog_notify/internal/service/core"
)

type Destination struct {
	botApi *tgbotapi.BotAPI

	markdownEscaper *strings.Replacer
}

func New(botToken string) (*Destination, error) {
	botApi, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}

	res := &Destination{
		botApi: botApi,
	}

	res.initMarkdownEscaper()

	return res, nil
}

func (s *Destination) Send(msg *core.Message) error {
	const chatId = -4624862121

	encodedMsg := s.EncodeMessage(msg)

	tgMsg := tgbotapi.NewMessage(chatId, encodedMsg)
	tgMsg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := s.botApi.Send(tgMsg)
	if err != nil {
		return fmt.Errorf("botApi.Send: %w", err)
	}

	return nil
}

func (s *Destination) EncodeMessage(m *core.Message) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf( // Заголовок сообщения
		" `%s` \\(*%s*\\):  %s\n\n",
		strings.ToUpper(s.escapeMarkdown(m.Tag)),
		strings.ToUpper(s.escapeMarkdown(m.Level)),
		s.escapeMarkdown(m.Ts),
	))
	if m.Message != "" {
		sb.WriteString(fmt.Sprintf("*•MESSAGE:* `%s`\n", s.escapeMarkdown(m.Message)))
	}
	if m.Error != "" {
		sb.WriteString(fmt.Sprintf("*•ERROR:* `%s`\n", s.escapeMarkdown(m.Error)))
	}

	// Обрабатываем остальные поля
	if len(m.V) > 0 {
		sb.WriteString(strings.Repeat("\\-", 30) + "\n")
		sortedKeys := m.GetVSortedKeys()

		for _, k := range sortedKeys {
			sb.WriteString(fmt.Sprintf("*•%s:* `%v`\n", s.escapeMarkdown(k), s.escapeMarkdown(fmt.Sprintf("%v", m.V[k]))))
		}
	}

	return sb.String()
}

func (s *Destination) initMarkdownEscaper() {
	s.markdownEscaper = strings.NewReplacer(
		`_`, `\_`,
		`-`, `\-`,
		`+`, `\+`,
		`.`, `\.`,
		`*`, `\*`,
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`~`, `\~`,
		`>`, `\>`,
		`#`, `\#`,
		`|`, `\|`,
	)
}

func (s *Destination) escapeMarkdown(v string) string {
	return s.markdownEscaper.Replace(v)
}
