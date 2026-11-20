package telegram

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/rendau/graylog_notify/internal/service/core"
)

const (
	defaultRuleKey = "default"

	// Telegram принимает до 4096 символов текста. Длину считаем по уже экранированной
	// строке: в ней символов не меньше, чем увидит Telegram, так что запас гарантирован.
	maxMessageLen = 4000
	// поля message и error бывают со стектрейсом на десятки килобайт
	maxMainFieldLen = 1500
	truncatedSuffix = "..."
)

type Destination struct {
	botApi          *tgbotapi.BotAPI
	rules           map[string][]int64
	defaultRule     []int64
	markdownEscaper *strings.Replacer
}

func New(apiUrl string, botToken string, inputRules []Rule) (*Destination, error) {
	var err error

	var botApi *tgbotapi.BotAPI

	if apiUrl != "" {
		apiUrl = strings.TrimRight(apiUrl, "/") + "/bot%s/%s"
		botApi, err = tgbotapi.NewBotAPIWithAPIEndpoint(botToken, apiUrl)
		if err != nil {
			return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
		}
	} else {
		botApi, err = tgbotapi.NewBotAPI(botToken)
		if err != nil {
			return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
		}
	}

	res := &Destination{
		botApi: botApi,
	}

	res.applyRules(inputRules)

	res.initMarkdownEscaper()

	slog.Info("Telegram default-rule", "chat_ids", res.defaultRule)
	for tag, chatIds := range res.rules {
		slog.Info("Telegram rule "+tag, "chat_ids", chatIds)
	}

	go res.reader()

	return res, nil
}

func (s *Destination) Send(msg *core.Message) error {
	chatIds := s.getChatIdsForTag(msg.Tag)
	if chatIds == nil {
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

	for _, chatId := range chatIds {
		tgMsg := tgbotapi.NewMessage(chatId, encodedMsg)
		tgMsg.ParseMode = tgbotapi.ModeMarkdownV2

		_, err := s.botApi.Send(tgMsg)
		if err != nil {
			return fmt.Errorf("botApi.Send: %w", err)
		}
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
		sb.WriteString(fmt.Sprintf("*•MESSAGE:* `%s`\n", s.escapeMarkdown(truncate(m.Message, maxMainFieldLen))))
	}
	if m.Error != "" {
		sb.WriteString(fmt.Sprintf("*•ERROR:* `%s`\n", s.escapeMarkdown(truncate(m.Error, maxMainFieldLen))))
	}

	// Обрабатываем остальные поля. Строки добавляем целиком, пока помещаются в лимит
	// Telegram: обрезать готовую разметку посередине нельзя — разорвётся `...`.
	if len(m.V) > 0 {
		sb.WriteString(strings.Repeat("\\-", 30) + "\n")
		sortedKeys := m.GetVSortedKeys()

		omitted := 0
		for _, k := range sortedKeys {
			line := fmt.Sprintf("*•%s:* `%s`\n", s.escapeMarkdown(k), s.escapeMarkdown(fmt.Sprintf("%v", m.V[k])))
			if omitted > 0 || utf8.RuneCountInString(sb.String())+utf8.RuneCountInString(line) > maxMessageLen-100 {
				omitted++
				continue
			}
			sb.WriteString(line)
		}
		if omitted > 0 {
			sb.WriteString(s.escapeMarkdown(fmt.Sprintf("... ещё полей не поместилось: %d", omitted)) + "\n")
		}
	}

	return sb.String()
}

// truncate обрезает строку до maxLen символов (рун), не разрывая многобайтные символы.
func truncate(v string, maxLen int) string {
	if utf8.RuneCountInString(v) <= maxLen {
		return v
	}

	return string([]rune(v)[:maxLen]) + truncatedSuffix
}

// initMarkdownEscaper — все символы, которые MarkdownV2 требует экранировать
// (https://core.telegram.org/bots/api#markdownv2-style). Экранированный символ Telegram
// считает обычным и внутри `кода`, поэтому один и тот же экранировщик годится везде.
// Replacer проходит строку один раз: вставленные им `\` повторно не экранируются.
func (s *Destination) initMarkdownEscaper() {
	s.markdownEscaper = strings.NewReplacer(
		`\`, `\\`,
		`_`, `\_`,
		`*`, `\*`,
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`~`, `\~`,
		"`", "\\`",
		`>`, `\>`,
		`#`, `\#`,
		`+`, `\+`,
		`-`, `\-`,
		`=`, `\=`,
		`|`, `\|`,
		`{`, `\{`,
		`}`, `\}`,
		`.`, `\.`,
		`!`, `\!`,
	)
}

func (s *Destination) escapeMarkdown(v string) string {
	return s.markdownEscaper.Replace(v)
}

func (s *Destination) applyRules(src []Rule) {
	s.rules = make(map[string][]int64, len(src))

	for _, ruleInput := range src {
		chatIds := s.normalizeChatIds(ruleInput.ChatIds)
		if len(chatIds) == 0 {
			continue
		}
		for _, tag := range ruleInput.Tags {
			tag = strings.ToLower(strings.TrimSpace(tag))
			if tag == "" {
				continue
			}
			if tag == defaultRuleKey {
				s.defaultRule = chatIds
			} else {
				s.rules[tag] = chatIds
			}
		}
	}
}

func (s *Destination) normalizeChatIds(v []int64) []int64 {
	result := make([]int64, 0, len(v))

	for _, chatId := range v {
		if chatId == 0 {
			continue
		}

		if !slices.Contains(result, chatId) {
			result = append(result, chatId)
		}
	}

	return result
}

func (s *Destination) getChatIdsForTag(tag string) []int64 {
	if chatIds, ok := s.rules[tag]; ok {
		return chatIds
	}

	return s.defaultRule
}

func (s *Destination) reader() {
	time.Sleep(time.Minute)

	for {
		s._reader()

		time.Sleep(time.Minute)
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
