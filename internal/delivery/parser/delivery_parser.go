package parser

import (
	"encoding/json"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-rabbitmq-lib/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

func ParseDeliveryJSON(out interface{}, delivery amqp091.Delivery) error {
	if delivery.ContentType != rabbitmq.CONTENT_TYPE_JSON {
		return errortypes.NewValidationException(
			fmt.Sprintf("ContentType (%s) should be %s",
				delivery.ContentType, rabbitmq.CONTENT_TYPE_JSON))
	}

	return json.Unmarshal(delivery.Body, out)
}
