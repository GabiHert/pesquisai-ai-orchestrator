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
	enumlanguages "github.com/PesquisAi/pesquisai-database-lib/sql/enums/languages"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"slices"
	"strings"
)

const (
	questionTemplate = `You are a part of a major project. In this project I will perform a google search, and your only` +
		` responsibility is to answer me, given the context of the pearson/company that are asking, the desired research` +
		` and the countries that will be used filter the results, what are the best languages that I should use to filter the Google search results. You should answer with a list of 2 digit ` +
		`language codes. Respond only with a comma separated list of language codes, nothing else. Consider that if the research will be filtered by the countries below, makes sense to match the country languages.` +
		` But any language that makes sense in the research context must be use.` +
		`Here I have a list of the codes you can use: %s. person/company context:"%s". research:"%s". countries:"%s".`
)

type languageService struct {
	queueGemini            interfaces.Queue
	queueOrchestrator      interfaces.Queue
	requestRepository      interfaces.RequestRepository
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l languageService) validateGeminiResponse(response []string) ([]string, []string) {
	var errorMessages, res []string
	for i, split := range response {
		if strings.Contains(split, "-") {
			split = strings.Split(split, "-")[0]
			response[i] = split
		}

		if !slices.Contains(enumlanguages.Languages, split) {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is not a valid language", split))
			continue
		}
		res = append(res, split)
	}

	if len(errorMessages) > 0 {
		return nil, errorMessages
	}

	return res, nil
}
func (l languageService) validateOrchestratorData(request nosqlmodels.Request) error {
	var messages []string
	if request.Context == nil {
		messages = append(messages, `"context" is required in mongoDB to perform language service`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required in mongoDB to perform language service`)
	}
	if request.Locations == nil {
		messages = append(messages, `"locations" is required in mongoDB to perform language service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l languageService) Execute(ctx context.Context, orchestratorRequest models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "languageService.Execute",
		slog.String("details", "process started"))

	var request nosqlmodels.Request
	err := l.orchestratorRepository.GetById(ctx, *orchestratorRequest.RequestId, &request)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateOrchestratorData(request)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question := l.buildQuestion(request)

	b, err := builder.BuildQueueGeminiMessage(
		*orchestratorRequest.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.Language,
		0,
	)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l languageService) buildQuestion(request nosqlmodels.Request) string {
	return fmt.Sprintf(
		questionTemplate,
		strings.Join(enumlanguages.Languages, ","),
		*request.Context,
		*request.Research,
		strings.Join(*request.Locations, ","),
	)
}

func (l languageService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "languageService.Callback",
		slog.String("details", "process started"))

	languages := strings.Split(strings.ToLower(*callback.Response), ",")
	languages, errMessages := l.validateGeminiResponse(languages)
	if errMessages != nil {
		var request nosqlmodels.Request
		err := l.orchestratorRepository.GetById(ctx, *callback.RequestId, &request)
		if err != nil {
			slog.ErrorContext(ctx, "languageService.Execute",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
		question := l.buildQuestion(request)
		err = errortypes.NewInvalidAIResponseException(*callback.RequestId, question, enumactions.Language, callback.ReceiveCount, errMessages...)
		slog.ErrorContext(ctx, "languageService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)
	for _, language := range languages {
		g.Go(func() error {
			e := l.requestRepository.RelateLanguage(groupCtx, *callback.RequestId, language)
			if e != nil && strings.Contains(e.Error(), `unique constraint "request_languages_pkey"`) {
				return nil
			}
			return e
		})
	}

	err := g.Wait()
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.orchestratorRepository.Update(ctx, *callback.RequestId,
		bson.M{"languages": languages},
	)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var (
		b      []byte
		action = enumactions.Sentences
	)
	b, err = builder.BuildQueueOrchestratorMessage(dtos.AiOrchestratorRequest{
		RequestId: callback.RequestId,
		Action:    &action,
	})
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueOrchestrator.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "languageService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	slog.InfoContext(ctx, "languageService.Callback",
		slog.String("details", "process finished"))
	return nil
}
func NewLanguageService(queueGemini, queueOrchestrator interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository, requestRepository interfaces.RequestRepository) interfaces.Service {
	return &languageService{
		queueGemini:            queueGemini,
		requestRepository:      requestRepository,
		orchestratorRepository: orchestratorRepository,
		queueOrchestrator:      queueOrchestrator,
	}
}
