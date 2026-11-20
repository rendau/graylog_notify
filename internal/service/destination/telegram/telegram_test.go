package telegram

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rendau/graylog_notify/internal/service/core"
)

// markdownV2Reserved — символы, которые Telegram требует экранировать в MarkdownV2.
const markdownV2Reserved = "_*[]()~`>#+-=|{}.!\\"

func newTestDestination() *Destination {
	d := &Destination{}
	d.initMarkdownEscaper()

	return d
}

// unescaped возвращает позиции символа ch, перед которым нет экранирующего `\`
// (чётное число обратных слэшей подряд — символ не экранирован).
func unescaped(s string, ch rune) []int {
	var result []int

	runes := []rune(s)
	for i, r := range runes {
		if r != ch {
			continue
		}
		slashes := 0
		for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
			slashes++
		}
		if slashes%2 == 0 {
			result = append(result, i)
		}
	}

	return result
}

func TestEscapeMarkdown_AllReserved(t *testing.T) {
	d := newTestDestination()

	for _, ch := range markdownV2Reserved {
		got := d.escapeMarkdown("a" + string(ch) + "b")
		assert.Equal(t, "a\\"+string(ch)+"b", got, "символ %q", ch)
	}

	// обычный текст и кириллица не трогаются
	assert.Equal(t, "Привет мир 123", d.escapeMarkdown("Привет мир 123"))
	// уже стоящий в тексте `\` экранируется ровно один раз
	assert.Equal(t, `C:\\path\=1`, d.escapeMarkdown(`C:\path=1`))
}

func TestEncodeMessage_ReservedCharsInValues(t *testing.T) {
	d := newTestDestination()

	// реальный случай: ошибка Telegram «Character '=' is reserved» — значения с = и {}
	msg := core.NewMessage("back", map[string]any{
		"level":   "error",
		"ts":      "2026-11-20T20:50:34+05:00",
		"msg":     "InfrastructureException: query a=1&b={2}!",
		"error":   "bad `quoted` value \\ end",
		"url":     "http://back-sps:443/api/v3/catalog/offers?productIds%5B0%5D=99a4937c&x=1",
		"context": `{"message":"Неверный запрос mindbox","line":131}`,
	})

	out := d.EncodeMessage(msg)

	// вне нашей разметки ни одного неэкранированного зарезервированного символа:
	// разметка — это `...` вокруг значений, *...* в заголовках полей и \( \) в шапке
	for _, ch := range "=!{}[]#+|~>_." {
		assert.Empty(t, unescaped(out, ch), "неэкранированный %q в:\n%s", ch, out)
	}
	// обратные кавычки парные: каждое значение открыто и закрыто
	assert.Equal(t, 0, len(unescaped(out, '`'))%2, out)
	assert.Contains(t, out, "Неверный запрос mindbox")
}

func TestEncodeMessage_LongMessageFitsTelegramLimit(t *testing.T) {
	d := newTestDestination()

	src := map[string]any{
		"level": "error",
		"msg":   strings.Repeat("Стектрейс = длинный! ", 2000),
		"error": strings.Repeat("e=", 5000),
	}
	for i := range 40 {
		src["field_"+strings.Repeat("x", i%5)+string(rune('a'+i%26))+string(rune('a'+i/26))] = strings.Repeat("значение.", 100)
	}

	out := d.EncodeMessage(core.NewMessage("back", src))

	require.LessOrEqual(t, utf8.RuneCountInString(out), 4096)
	assert.Equal(t, 0, len(unescaped(out, '`'))%2, "разметка не разорвана")
	assert.Contains(t, out, "не поместилось")
	assert.True(t, utf8.ValidString(out))
}
