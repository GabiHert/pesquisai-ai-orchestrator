package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/dtos"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	enumactions "github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/enums/actions"
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
	locationQuestionTemplate = `You are a part of a major project. In this project I will perform a google search, and your only` +
		` responsibility is to answer me, given the context of the pearson/company that are asking and the desired research,` +
		` what are the best countries that I should filter the Google search results. You should answer with a list of 2 digit ` +
		`country codes. Respond only with a comma separated list of country codes, nothing else. ` +
		`Here I have a list of the codes you can use: %s. person/company context:"%s". research:"%s".`
)

type locationService struct {
	queueGemini            interfaces.Queue
	queueOrchestrator      interfaces.Queue
	requestRepository      interfaces.RequestRepository
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l locationService) validateGeminiResponse(response []string) *string {
	for _, split := range response {
		if !slices.Contains(enumlocations.Locations, split) {
			message := fmt.Sprintf("%s is not a valid location", split)
			return &message
		}
	}
	return nil
}

func (l locationService) validateRequest(request models.AiOrchestratorRequest) error {
	var messages []string
	if request.Context == nil {
		messages = append(messages, `"context" is required`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l locationService) buildQuestion(context, research string) string {
	return fmt.Sprintf(
		locationQuestionTemplate,
		strings.Join(enumlocations.Locations, ","),
		context,
		research)
}

func (l locationService) Execute(ctx context.Context, request models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "locationService.Execute",
		slog.String("details", "process started"))

	err := l.validateRequest(request)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	createdAt := time.Now().UTC()
	err = l.orchestratorRepository.Create(ctx, &nosqlmodels.Request{
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

	question := l.buildQuestion(*request.Context, *request.Research)

	b, err := builder.BuildQueueGeminiMessage(
		*request.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.Location,
		0,
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

func (l locationService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "locationService.Callback",
		slog.String("details", "process started"))

	locations := strings.Split(strings.ToLower(*callback.Response), ",")
	errMessage := l.validateGeminiResponse(locations)
	if errMessage != nil {
		var request nosqlmodels.Request
		err := l.orchestratorRepository.GetById(ctx, *callback.RequestId, &request)
		if err != nil {
			slog.ErrorContext(ctx, "locationService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
		question := l.buildQuestion(*request.Context, *request.Research)
		err = errortypes.NewInvalidAIResponseException(*request.ID, question, enumactions.Location, callback.ReceiveCount+1, *errMessage)
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)
	for _, location := range locations {
		g.Go(func() error {
			e := l.requestRepository.RelateLocation(groupCtx, *callback.RequestId, location)
			if e != nil && strings.Contains(e.Error(), `unique constraint "request_locations_pkey"`) {
				return nil
			}
			return e
		})
	}

	err := g.Wait()
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.orchestratorRepository.Update(ctx, *callback.RequestId,
		bson.M{"locations": locations},
	)
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var (
		b      []byte
		action = enumactions.Language
	)
	b, err = builder.BuildQueueOrchestratorMessage(dtos.AiOrchestratorRequest{
		RequestId: callback.RequestId,
		Action:    &action,
	})
	if err != nil {
		slog.ErrorContext(ctx, "locationService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueOrchestrator.Publish(ctx, b)
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

func NewLocationService(queueGemini, queueOrchestrator interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository, requestRepository interfaces.RequestRepository) interfaces.Service {
	return &locationService{
		queueGemini:            queueGemini,
		requestRepository:      requestRepository,
		orchestratorRepository: orchestratorRepository,
		queueOrchestrator:      queueOrchestrator,
	}
}
