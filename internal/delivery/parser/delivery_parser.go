package parser

import (
	"encoding/json"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-rabbitmq-lib/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
)

func ParseDeliveryJSON(out interface{}, delivery amqp091.Delivery) error {
	if delivery.ContentType != rabbitmq.ContentTypeJson {
		return errortypes.NewValidationException(
			fmt.Sprintf("ContentType (%s) should be %s",
				delivery.ContentType, rabbitmq.ContentTypeJson))
	}

	return json.Unmarshal(delivery.Body, out)
}
