package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-api/internal/config/properties"
	"github.com/PesquisAi/pesquisai-api/internal/domain/builder"
	"github.com/PesquisAi/pesquisai-api/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-api/internal/domain/models"
	"log/slog"
)

const (
	questionTemplate = "%s %s"
)

type locationService struct {
	queueGemini interfaces.Queue
}

func (l locationService) Execute(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "locationService.Execute",
		slog.String("details", "process started"))

	forward := map[string]any{
		"request_id": *request.RequestId,
		"context":    *request.Context,
		"action":     *request.Action,
		"research":   *request.Research,
	}

	question := fmt.Sprintf(
		questionTemplate,
		*request.Context,
		*request.Research)

	b, err := builder.BuildQueueGeminiMessage(
		forward,
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
