package usecases

import (
	"context"
	"github.com/PesquisAi/pesquisai-api/internal/domain/factory"
	"github.com/PesquisAi/pesquisai-api/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-api/internal/domain/models"
	"log/slog"
)

type UseCase struct {
	requestRepository interfaces.RequestRepository
	serviceFactory    *factory.ServiceFactory
}

func (u UseCase) Orchestrate(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "useCase.Orchestrate",
		slog.String("details", "process started"))

	service, err := u.serviceFactory.Factory(request)
	if err != nil {
		slog.ErrorContext(ctx, "useCase.Orchestrate",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = service.Execute(ctx, request)
	if err != nil {
		slog.ErrorContext(ctx, "useCase.Orchestrate",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	slog.DebugContext(ctx, "useCase.Orchestrate",
		slog.String("details", "process finished"))
	return nil
}

func NewUseCase(requestRepository interfaces.RequestRepository, serviceFactory *factory.ServiceFactory) interfaces.UseCase {
	return &UseCase{
		requestRepository: requestRepository,
		serviceFactory:    serviceFactory,
	}
}
