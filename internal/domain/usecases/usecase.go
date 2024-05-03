package usecases

import (
	"context"
	"github.com/GabiHert/pesquisai-api/internal/domain/factory"
	"github.com/GabiHert/pesquisai-api/internal/domain/interfaces"
	"github.com/GabiHert/pesquisai-api/internal/domain/models"
	"log/slog"
)

type UseCase struct {
	requestRepository interfaces.RequestRepository
	queueGemini       interfaces.Queue
}

func (u UseCase) Orchestrate(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "useCase.Orchestrate",
		slog.String("details", "process started"))

	service, err := factory.FactorService(request)
	if err != nil {
		return err
	}

	err = service.Execute()
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "useCase.Orchestrate",
		slog.String("details", "process finished"))
	return nil
}

func NewUseCase(requestRepository interfaces.RequestRepository, aiOrchestratorQueue interfaces.Queue) interfaces.UseCase {
	return &UseCase{
		requestRepository: requestRepository,
	}
}
