package controllers

import (
	"context"
	"errors"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/dtos"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/parser"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/validations"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	"github.com/PesquisAi/pesquisai-errors-lib/exceptions"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type controller struct {
	useCase     interfaces.UseCase
	queueGemini interfaces.Queue
}

func (c controller) errorHandler(ctx context.Context, err error) error {
	var exception *exceptions.Error
	if !errors.As(err, &exception) {
		exception = errortypes.NewUnknownException(err.Error())
	}

	b, _ := exception.ToJSON()
	slog.Error("controller.errorHandler",
		slog.String("details", "process error"),
		slog.String("errorType", string(b)))

	if exception.Code == errortypes.InvalidAiResponseCode {
		receiveCount, _ := exception.Forward["receiveCount"].(int)
		if receiveCount >= properties.GetMaxAiReceiveCount() {
			return nil
		}

		requestId, _ := exception.Forward["requestId"].(string)
		question, _ := exception.Forward["question"].(string)
		action, _ := exception.Forward["action"].(string)

		b, err = builder.BuildQueueGeminiMessage(requestId, question, properties.QueueNameAiOrchestratorCallback, action, receiveCount)
		if err != nil {
			slog.Error("controller.errorHandler",
				slog.String("details", "process error"),
				slog.Any("err", err.Error()))
		} else {
			err = c.queueGemini.Publish(ctx, b)
			if err == nil {
				return nil
			}
			slog.Warn("controller.errorHandler",
				slog.String("details", "process error"),
				slog.Any("err", err.Error()))
		}
	}

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
		_ = c.errorHandler(context.TODO(), err)
	}
}

func (c controller) AiOrchestratorHandler(delivery amqp.Delivery) error {
	defer c.def()
	ctx := context.Background()
	slog.Info("controller.AiOrchestratorHandler",
		slog.String("details", "process started"),
		slog.String("messageId", delivery.MessageId),
		slog.String("userId", delivery.UserId))

	var request dtos.AiOrchestratorRequest
	err := parser.ParseDeliveryJSON(&request, delivery)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	err = validations.ValidateRequest(&request)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	requestModel := models.AiOrchestratorRequest{
		RequestId:  request.RequestId,
		ResearchId: request.ResearchId,
		Context:    request.Context,
		Research:   request.Research,
		Action:     request.Action,
	}

	err = c.useCase.Orchestrate(ctx, requestModel)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	slog.Info("controller.AiOrchestratorHandler",
		slog.String("details", "process finished"))
	return nil
}

func (c controller) AiOrchestratorCallbackHandler(delivery amqp.Delivery) error {
	defer c.def()
	ctx := context.Background()

	slog.Info("controller.AiOrchestratorCallbackHandler",
		slog.String("details", "process started"),
		slog.String("messageId", delivery.MessageId),
		slog.String("userId", delivery.UserId))

	var callback dtos.AiOrchestratorCallbackRequest
	err := parser.ParseDeliveryJSON(&callback, delivery)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	err = validations.ValidateCallbackRequest(&callback)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	requestModel := models.AiOrchestratorCallbackRequest{
		RequestId:    callback.RequestId,
		ResearchId:   callback.ResearchId,
		Response:     callback.Response,
		Action:       callback.Forward.Action,
		ReceiveCount: *callback.Forward.ReceiveCount,
	}

	err = c.useCase.OrchestrateCallback(ctx, requestModel)
	if err != nil {
		return c.errorHandler(ctx, err)
	}

	slog.Info("controller.AiOrchestratorCallbackHandler",
		slog.String("details", "process finished"))

	return nil
}

func NewController(queueGemini interfaces.Queue, useCase interfaces.UseCase) interfaces.Controller {
	return &controller{
		useCase:     useCase,
		queueGemini: queueGemini,
	}
}
