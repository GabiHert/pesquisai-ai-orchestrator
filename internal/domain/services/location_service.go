package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	nosqlmodels "github.com/PesquisAi/pesquisai-database-lib/nosql/models"
	enumlocations "github.com/PesquisAi/pesquisai-database-lib/sql/enums/locations"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"slices"
	"strings"
	"time"
)

const (
	questionTemplate = `You are a part of a major project. In this project I will perform a google search, and your only` +
		` responsibility is to answer me, given the context of the pearson/company that are asking and the desired research,` +
		` what are the best countries that I should filter the Google search results. You should answer with a list of 2 digit ` +
		`country codes. Respond only with a comma separated list of country codes, nothing else. ` +
		`Here I have a list of the codes you can use: %s. context:"%s". research:"%s".`
)

type locationService struct {
	queueGemini            interfaces.Queue
	requestRepository      interfaces.RequestRepository
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l locationService) validateResponse(response []string) error {
	for _, split := range response {
		if !slices.Contains(enumlocations.Locations, split) {
			return errortypes.NewValidationException(fmt.Sprintf("%s is not a valid location", split))
		}
	}
	return nil
}

func (l locationService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "locationService.Callback",
		slog.String("details", "process started"))

	languages := strings.Split(*callback.Response, ",")
	err := l.validateResponse(languages)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)
	for _, language := range languages {
		g.Go(func() error {
			//todo: relate language
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

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
