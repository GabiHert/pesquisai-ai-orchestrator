package interfaces

import (
	"context"
	"github.com/GabiHert/pesquisai-api/internal/domain/models"
)

type UseCase interface {
	Orchestrate(ctx context.Context, request models.AiOrchestratorRequest) error
}
