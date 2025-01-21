package telegram

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rendau/graylog_notify/internal/service/core"
)

const (
	defaultRuleKey = "default"
)

type Destination struct {
	botApi          *tgbotapi.BotAPI
	rules           map[string]*Rule
	defaultRule     *Rule
	markdownEscaper *strings.Replacer
}

func New(botToken string, rules map[string]*Rule) (*Destination, error) {
	botApi, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}

	res := &Destination{
		botApi: botApi,
	}

	res.applyRules(rules)

	res.initMarkdownEscaper()

	slog.Info("Telegram started with rules", "rules", res.rules, "default_rule", res.defaultRule)

	go res.reader()

	return res, nil
}

func (s *Destination) Send(msg *core.Message) error {
	rule := s.findRuleByTag(msg.Tag)
	if rule == nil {
		slog.Info(
			"Tg-message, no rule found for tag "+msg.Tag,
			"tag", msg.Tag,
			"level", msg.Level,
			"ts", msg.Ts,
			"message", msg.Message,
			"error", msg.Error,
			"other_fields", msg.V,
		)
		return nil
	}

	encodedMsg := s.EncodeMessage(msg)

	tgMsg := tgbotapi.NewMessage(rule.ChatId, encodedMsg)
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

func (s *Destination) applyRules(src map[string]*Rule) {
	s.rules = make(map[string]*Rule, len(src))

	for k, v := range src {
		if v == nil || v.ChatId == 0 {
			continue
		}
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" {
			continue
		}
		if k == defaultRuleKey {
			s.defaultRule = v
		} else {
			s.rules[k] = v
		}
	}
}

func (s *Destination) findRuleByTag(tag string) *Rule {
	if rule, ok := s.rules[tag]; ok {
		return rule
	}

	return s.defaultRule
}

func (s *Destination) reader() {
	for {
		s._reader()

		time.Sleep(time.Second * 30)
	}
}

func (s *Destination) _reader() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Tg-reader recovered from panic", slog.Any("error", err))
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.botApi.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		logArgs := make([]any, 0, 20)

		if update.Message != nil {
			if update.Message.From != nil {
				logArgs = append(logArgs, "from_id", update.Message.From.ID)
				logArgs = append(logArgs, "from_user_name", update.Message.From.UserName)
			}

			if update.Message.Chat != nil {
				logArgs = append(logArgs, "chat_id", update.Message.Chat.ID)
			}

			if update.Message.MigrateFromChatID != 0 {
				logArgs = append(logArgs, "migrate_from_chat_id", update.Message.MigrateFromChatID)
			}

			if update.Message.MigrateToChatID != 0 {
				logArgs = append(logArgs, "migrate_to_chat_id", update.Message.MigrateToChatID)
			}

			if update.Message.Text != "" {
				logArgs = append(logArgs, "message_text", update.Message.Text)
			}
		}

		slog.Info("Tg-message", logArgs...)
	}
}
