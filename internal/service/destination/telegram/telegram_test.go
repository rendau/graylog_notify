package telegram

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rendau/graylog_notify/internal/service/core"
)

func TestDestination_Send(t *testing.T) {
	const botToken = "1843316279:AAEyleWu4EYbkkuxjAIlLlwzXzY9BHPhZhY"

	tg, err := New(botToken)
	require.NoError(t, err)

	msgSrc := map[string]any{
		"level":    "error",
		"ts":       "2023-11-01T14:54:41+06:00",
		"caller":   "httpc/struct.go:187",
		"msg":      "FF: LinkOrderCondition: Bad status code",
		"error":    "bad_status_code",
		"method":   "PUT",
		"uri":      "https://api.ffin.credit/ffc-api-public/universal/general/set-reference-id/2007a855-b84b-4840-82ad-f01ff7d2c88c",
		"params":   "",
		"req_body": "{\"reference_id\":\"841145\",\"credit_params\":{\"period\":12,\"principal\":108820},\"product\":\"MECHTA_CREDIT12\"}",
	}

	msg := core.NewMessage("credit-broker", msgSrc)

	err = tg.Send(msg)
	require.NoError(t, err)
}
