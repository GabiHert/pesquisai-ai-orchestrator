package interfaces

import (
	"context"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
)

type UseCase interface {
	Orchestrate(ctx context.Context, request models.AiOrchestratorRequest) error
}
