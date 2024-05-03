package services

import (
	"context"
	"fmt"
	"github.com/GabiHert/pesquisai-api/internal/config/properties"
	"github.com/GabiHert/pesquisai-api/internal/domain/builder"
	"github.com/GabiHert/pesquisai-api/internal/domain/interfaces"
	"github.com/GabiHert/pesquisai-api/internal/domain/models"
)

const (
	questionTemplate = "%s %s"
)

type locationService struct {
	request     models.AiOrchestratorRequest
	queueGemini interfaces.Queue
}

func (l locationService) Execute(ctx context.Context) error {
	forward := map[string]any{
		"request_id": *l.request.RequestId,
		"context":    *l.request.Context,
		"action":     *l.request.Action,
		"research":   *l.request.Research,
	}

	question := fmt.Sprintf(
		questionTemplate,
		*l.request.Context,
		*l.request.Research)

	b, err := builder.BuildQueueGeminiMessage(
		forward,
		*l.request.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
	)
	if err != nil {
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	return err
}

func NewLocationService(request models.AiOrchestratorRequest, queueGemini interfaces.Queue) interfaces.Service {
	return &locationService{request: request, queueGemini: queueGemini}
}
