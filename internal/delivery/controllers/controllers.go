package controllers

import (
	"context"
	"github.com/PesquisAi/pesquisai-api/internal/delivery/dtos"
	"github.com/PesquisAi/pesquisai-api/internal/delivery/parser"
	"github.com/PesquisAi/pesquisai-api/internal/delivery/validations"
	"github.com/PesquisAi/pesquisai-api/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-api/internal/domain/models"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type controller struct {
	useCase interfaces.UseCase
}

func (c controller) errorHandler(err error) {}

func (c controller) AiOrchestratorHandler(delivery amqp.Delivery) {
	slog.Info("useCase.Create",
		slog.String("details", "process started"),
		slog.String("messageId", delivery.MessageId),
		slog.String("userId", delivery.UserId))

	var request dtos.AiOrchestratorRequest
	err := parser.ParseDeliveryJSON(&request, delivery)
	if err != nil {
		c.errorHandler(err)
		return
	}

	err = validations.Validate(&request)
	if err != nil {
		c.errorHandler(err)
		return
	}

	requestModel := models.AiOrchestratorRequest{
		RequestId: request.RequestId,
		Context:   request.Context,
		Research:  request.Research,
		Action:    request.Action,
	}

	err = c.useCase.Orchestrate(context.Background(), requestModel)
	if err != nil {
		c.errorHandler(err)
		return
	}

	slog.Info("useCase.Create",
		slog.String("details", "process finished"))
}

func (c controller) AiOrchestratorCallbackHandler(delivery amqp.Delivery) {
	//TODO implement me
	panic("implement me")
}

func NewController(useCase interfaces.UseCase) interfaces.Controller {
	return &controller{useCase}
}
