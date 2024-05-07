package controllers

import (
	"context"
	"errors"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/dtos"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/parser"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/validations"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	"github.com/PesquisAi/pesquisai-errors-lib/exceptions"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type controller struct {
	useCase interfaces.UseCase
}

func (c controller) errorHandler(err error) {
	exception := &exceptions.Error{}
	if !errors.As(err, &exception) {
		exception = errortypes.NewUnknownException(err.Error())
	}

	b, _ := exception.ToJSON()
	slog.Error("controller.errorHandler",
		slog.String("details", "process error"),
		slog.String("errorType", string(b)))
}

func (c controller) def() {

}

func (c controller) AiOrchestratorHandler(delivery amqp.Delivery) {
	defer c.def()
	slog.Info("controller.AiOrchestratorHandler",
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

	slog.Info("controller.AiOrchestratorHandler",
		slog.String("details", "process finished"))
	err = delivery.Ack(false)
	if err != nil {
		c.errorHandler(err)
		return
	}
}

func (c controller) AiOrchestratorCallbackHandler(delivery amqp.Delivery) {
	//TODO implement me
	panic("implement me")
}

func NewController(useCase interfaces.UseCase) interfaces.Controller {
	return &controller{useCase}
}
