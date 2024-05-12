package interfaces

import (
	"context"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
)

type Service interface {
	Execute(ctx context.Context, request models.AiOrchestratorRequest) error
	Callback(ctx context.Context, request models.AiOrchestratorCallbackRequest) error
}
