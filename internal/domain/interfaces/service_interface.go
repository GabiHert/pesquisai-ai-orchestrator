package interfaces

import (
	"context"
	"github.com/PesquisAi/pesquisai-api/internal/domain/models"
)

type Service interface {
	Execute(ctx context.Context, request models.AiOrchestratorRequest) error
}
