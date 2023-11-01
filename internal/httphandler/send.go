package httphandler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/rendau/graylog_notify/internal/core"
)

/*
{
  "event_definition_id": "NotificationTestId",
  "event_definition_type": "test-dummy-v1",
  "event_definition_title": "Event Definition Test Title",
  "event_definition_description": "Event Definition Test Description",
  "job_definition_id": "<unknown>",
  "job_trigger_id": "<unknown>",
  "event": {
    "id": "TEST_NOTIFICATION_ID",
    "event_definition_type": "notification-test-v1",
    "event_definition_id": "EventDefinitionTestId",
    "origin_context": "urn:graylog:message:es:testIndex_42:b5e53442-12bb-4374-90ed-0deadbeefbaz",
    "timestamp": "2023-11-01T08:20:00.644Z",
    "timestamp_processing": "2023-11-01T08:20:00.644Z",
    "timerange_start": null,
    "timerange_end": null,
    "streams": [
      "000000000000000000000002"
    ],
    "source_streams": [],
    "message": "Notification test message triggered from user <local:admin>",
    "source": "000000000000000000000001",
    "key_tuple": [
      "testkey"
    ],
    "key": "testkey",
    "priority": 2,
    "alert": true,
    "fields": {
      "field1": "value1",
      "field2": "value2"
    },
    "group_by_fields": {}
  },
  "backlog": []
}
*/

/*
{
  "event_definition_id": "65420b968189520c8fe81364",
  "event_definition_type": "aggregation-v1",
  "event_definition_title": "Error and Warn",
  "event_definition_description": "",
  "job_definition_id": "65420a858189520c8fe81123",
  "job_trigger_id": "65420cc38189520c8fe815e2",
  "event": {
    "id": "01HE503PM994BYSAWWWEDD5CTZ",
    "event_definition_type": "aggregation-v1",
    "event_definition_id": "65420b968189520c8fe81364",
    "origin_context": "urn:graylog:message:es:graylog_0:9c63bd30-7890-11ee-beae-0242ac120004",
    "timestamp": "2023-11-01T08:28:18.943Z",
    "timestamp_processing": "2023-11-01T08:30:59.209Z",
    "timerange_start": null,
    "timerange_end": null,
    "streams": [],
    "source_streams": [
      "000000000000000000000001"
    ],
    "message": "Error and Warn",
    "source": "497115bebf4d",
    "key_tuple": [],
    "key": "",
    "priority": 3,
    "alert": true,
    "fields": {},
    "group_by_fields": {}
  },
  "backlog": []
}

{
  "event_definition_id": "65420b968189520c8fe81364",
  "event_definition_type": "aggregation-v1",
  "event_definition_title": "Error and Warn",
  "event_definition_description": "",
  "job_definition_id": "65420a858189520c8fe81123",
  "job_trigger_id": "654212a48189520c8fe8222c",
  "event": {
    "id": "01HE51HMFN00N59KTY6QSBPBAK",
    "event_definition_type": "aggregation-v1",
    "event_definition_id": "65420b968189520c8fe81364",
    "origin_context": "urn:graylog:message:es:graylog_0:4b984f71-7894-11ee-beae-0242ac120004",
    "timestamp": "2023-11-01T08:54:41.382Z",
    "timestamp_processing": "2023-11-01T08:56:04.341Z",
    "timerange_start": null,
    "timerange_end": null,
    "streams": [],
    "source_streams": [
      "000000000000000000000001"
    ],
    "message": "Error and Warn",
    "source": "497115bebf4d",
    "key_tuple": [],
    "key": "",
    "priority": 3,
    "alert": true,
    "fields": {
      "tag": "credit-broker",
      "message": "{\"level\":\"error\",\"ts\":\"2023-11-01T14:54:41+06:00\",\"caller\":\"httpc/struct.go:187\",\"msg\":\"FF: LinkOrderCondition: Bad status code\",\"error\":\"bad_status_code\",\"method\":\"PUT\",\"uri\":\"https://api.ffin.credit/ffc-api-public/universal/general/set-reference-id/2007a855-b84b-4840-82ad-f01ff7d2c88c\",\"params\":\"\",\"req_body\":\"{\\\"reference_id\\\":\\\"841145\\\",\\\"credit_params\\\":{\\\"period\\\":12,\\\"principal\\\":108820},\\\"product\\\":\\\"MECHTA_CREDIT12\\\"}\",\"status_code\":400,\"rep_body\":\"[\\\"principal not in Product principal_limits\\\"]\",\"stacktrace\":\"github.com/mechta-market/dop/adapters/client/httpc.(*RespSt).LogError\\n\\t/home/runner/go/pkg/mod/github.com/mechta-market/dop@v1.0.10/adapters/client/httpc/struct.go:187\\ngithub.com/mechta-market/credit-broker/internal/adapters/provider/ff.(*St).sendRequest\\n\\t/home/runner/work/credit-broker/credit-broker/internal/adapters/provider/ff/index.go:497\\ngithub.com/mechta-market/credit-broker/internal/adapters/provider/ff.(*St).SendClientChoice\\n\\t/home/runner/work/credit-broker/credit-broker/internal/adapters/provider/ff/index.go:200\\ngithub.com/mechta-market/credit-broker/internal/domain/core.(*Ord).SendClientChoice\\n\\t/home/runner/work/credit-broker/credit-broker/internal/domain/core/ord.go:759\\ngithub.com/mechta-market/credit-broker/internal/domain/core.(*Ord).CreateContract\\n\\t/home/runner/work/credit-broker/credit-broker/internal/domain/core/ord.go:724\\ngithub.com/mechta-market/credit-broker/internal/domain/usecases.(*St).OrdCreateContract.func1\\n\\t/home/runner/work/credit-broker/credit-broker/internal/domain/usecases/ord.go:131\\ngithub.com/mechta-market/dop/adapters/db/pg.(*St).TransactionFn\\n\\t/home/runner/go/pkg/mod/github.com/mechta-market/dop@v1.0.10/adapters/db/pg/index.go:158\\ngithub.com/mechta-market/credit-broker/internal/domain/usecases.(*St).OrdCreateContract\\n\\t/home/runner/work/credit-broker/credit-broker/internal/domain/usecases/ord.go:130\\ngithub.com/mechta-market/credit-broker/internal/adapters/server/rest.(*St).hOrdCreateContract\\n\\t/home/runner/work/credit-broker/credit-broker/internal/adapters/server/rest/h_ord.go:187\\ngithub.com/gin-gonic/gin.(*Context).Next\\n\\t/home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:174\\ngithub.com/mechta-market/credit-broker/internal/adapters/server/rest.GetHandler.MwRecovery.func3\\n\\t/home/runner/go/pkg/mod/github.com/mechta-market/dop@v1.0.10/adapters/server/https/index.go:185\\ngithub.com/gin-gonic/gin.(*Context).Next\\n\\t/home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:174\\ngithub.com/gin-gonic/gin.(*Engine).handleHTTPRequest\\n\\t/home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/gin.go:620\\ngithub.com/gin-gonic/gin.(*Engine).ServeHTTP\\n\\t/home/runner/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/gin.go:576\\nnet/http.serverHandler.ServeHTTP\\n\\t/opt/hostedtoolcache/go/1.21.3/x64/src/net/http/server.go:2938\\nnet/http.(*conn).serve\\n\\t/opt/hostedtoolcache/go/1.21.3/x64/src/net/http/server.go:2009\"}"
    },
    "group_by_fields": {}
  },
  "backlog": []
}

{
  "event_definition_id": "65420b968189520c8fe81364",
  "event_definition_type": "aggregation-v1",
  "event_definition_title": "Error and Warn",
  "event_definition_description": "",
  "job_definition_id": "65420a858189520c8fe81123",
  "job_trigger_id": "654213cb8189520c8fe82498",
  "event": {
    "id": "01HE51TMQT0G3V67XKXRBQP05W",
    "event_definition_type": "aggregation-v1",
    "event_definition_id": "65420b968189520c8fe81364",
    "origin_context": "urn:graylog:message:es:graylog_0:1ec48581-7895-11ee-beae-0242ac120004",
    "timestamp": "2023-11-01T09:00:35.664Z",
    "timestamp_processing": "2023-11-01T09:00:59.514Z",
    "timerange_start": null,
    "timerange_end": null,
    "streams": [],
    "source_streams": [
      "000000000000000000000001"
    ],
    "message": "Error and Warn",
    "source": "497115bebf4d",
    "key_tuple": [],
    "key": "",
    "priority": 3,
    "alert": true,
    "fields": {
      "tag": "stg",
      "message": "{\"level\":\"warn\",\"ts\":\"2023-11-01T15:00:35+06:00\",\"caller\":\"rest/index.go:47\",\"msg\":\"Warn message\",\"foo\":\"bar\",\"hello\":123}"
    },
    "group_by_fields": {}
  },
  "backlog": []
}
*/

type msgSt struct {
	Event struct {
		Fields map[string]string `json:"fields"`
	} `json:"event"`
}

func (h *handlerSt) Send(w http.ResponseWriter, r *http.Request) {
	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("fail to read body", slog.String("error", err.Error()))
		return
	}

	msg := msgSt{}
	err = json.Unmarshal(body, &msg)
	if err != nil {
		slog.Error(
			"fail to unmarshal body",
			slog.String("error", err.Error()),
			slog.String("body", string(body)),
		)
		return
	}
	if msg.Event.Fields == nil {
		return
	}

	tag := msg.Event.Fields["tag"]
	message := msg.Event.Fields["message"]

	if message == "" {
		return
	}

	h.cr.Send(core.SendRequestSt{
		Tag:     tag,
		Message: message,
	})
}
