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

func (c controller) errorHandler(err error) error {
	exception := &exceptions.Error{}
	if !errors.As(err, &exception) {
		exception = errortypes.NewUnknownException(err.Error())
	}

	b, _ := exception.ToJSON()
	slog.Error("controller.errorHandler",
		slog.String("details", "process error"),
		slog.String("errorType", string(b)))

	if exception.Abort {
		return nil
	}
	return exception
}

func (c controller) def() {
	if r := recover(); r != nil {
		slog.Error("controller.def",
			slog.String("details", "process panic"),
			slog.Any("recover", r))

		err := errortypes.NewUnknownException("application panic")
		_ = c.errorHandler(err)
	}
}

func (c controller) AiOrchestratorHandler(delivery amqp.Delivery) error {
	defer c.def()
	slog.Info("controller.AiOrchestratorHandler",
		slog.String("details", "process started"),
		slog.String("messageId", delivery.MessageId),
		slog.String("userId", delivery.UserId))

	var request dtos.AiOrchestratorRequest
	err := parser.ParseDeliveryJSON(&request, delivery)
	if err != nil {
		return c.errorHandler(err)
	}

	err = validations.ValidateRequest(&request)
	if err != nil {
		return c.errorHandler(err)
	}

	requestModel := models.AiOrchestratorRequest{
		RequestId: request.RequestId,
		Context:   request.Context,
		Research:  request.Research,
		Action:    request.Action,
	}

	err = c.useCase.Orchestrate(context.Background(), requestModel)
	if err != nil {
		return c.errorHandler(err)
	}

	slog.Info("controller.AiOrchestratorHandler",
		slog.String("details", "process finished"))
	return nil
}

func (c controller) AiOrchestratorCallbackHandler(delivery amqp.Delivery) error {
	defer c.def()
	slog.Info("controller.AiOrchestratorCallbackHandler",
		slog.String("details", "process started"),
		slog.String("messageId", delivery.MessageId),
		slog.String("userId", delivery.UserId))

	var callback dtos.AiOrchestratorCallbackRequest
	err := parser.ParseDeliveryJSON(&callback, delivery)
	if err != nil {
		return c.errorHandler(err)
	}

	err = validations.ValidateCallbackRequest(&callback)
	if err != nil {
		return c.errorHandler(err)
	}

	requestModel := models.AiOrchestratorCallbackRequest{
		RequestId:  callback.RequestId,
		ResearchId: callback.ResearchId,
		Response:   callback.Response,
		Action:     callback.Forward.Action,
	}

	err = c.useCase.OrchestrateCallback(context.Background(), requestModel)
	if err != nil {
		return c.errorHandler(err)
	}

	slog.Info("controller.AiOrchestratorCallbackHandler",
		slog.String("details", "process finished"))

	return nil
}

func NewController(useCase interfaces.UseCase) interfaces.Controller {
	return &controller{useCase}
}
