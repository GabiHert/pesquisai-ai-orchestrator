package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	"log/slog"
)

const (
	questionTemplate = "%s %s"
)

type locationService struct {
	queueGemini       interfaces.Queue
	requestRepository interfaces.RequestRepository
}

func (l locationService) Execute(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "locationService.Execute",
		slog.String("details", "process started"))

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
