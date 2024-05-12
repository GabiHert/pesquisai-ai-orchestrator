package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	nosqlmodels "github.com/PesquisAi/pesquisai-database-lib/nosql/models"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
	"time"
)

const (
	questionTemplate = "%s %s"
)

type locationService struct {
	queueGemini            interfaces.Queue
	requestRepository      interfaces.RequestRepository
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l locationService) validateResponse(response string) error {

	return nil
}

func (l locationService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "locationService.Callback",
		slog.String("details", "process started"))

	err := l.validateResponse(*callback.Response)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var languages []string

	//todo: create relation request - language in database for each language

	err = l.orchestratorRepository.Update(ctx, *callback.RequestId,
		bson.M{"languages": languages},
	)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	slog.InfoContext(ctx, "locationService.Callback",
		slog.String("details", "process finished"))
	return nil
}

func (l locationService) Execute(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "locationService.Execute",
		slog.String("details", "process started"))

	createdAt := time.Now().UTC()
	err := l.orchestratorRepository.Create(ctx, nosqlmodels.Request{
		ID:        request.RequestId,
		Context:   request.Context,
		Research:  request.Research,
		CreatedAt: &createdAt,
		UpdatedAt: &createdAt,
	})
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question := fmt.Sprintf(
		questionTemplate,
		*request.Context,
		*request.Research)

	b, err := builder.BuildQueueGeminiMessage(
		*request.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
	)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func NewLocationService(queueGemini interfaces.Queue) interfaces.Service {
	return &locationService{queueGemini: queueGemini}
}
