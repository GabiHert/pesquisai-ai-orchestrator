package interfaces

import amqp "github.com/rabbitmq/amqp091-go"

type Controller interface {
	AiOrchestratorHandler(delivery amqp.Delivery)
	AiOrchestratorCallbackHandler(delivery amqp.Delivery)
}
