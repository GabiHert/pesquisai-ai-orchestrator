package services

import (
	"context"
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/builder"
	enumactions "github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/enums/actions"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
	nosqlmodels "github.com/PesquisAi/pesquisai-database-lib/nosql/models"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
	"strings"
)

const (
	sentenceQuestionTemplate = `You are a part of a major project. In this project I will perform a google search, and your only` +
		` responsibility is to answer me, given the context of the pearson/company that are asking and the research they want to do` +
		`, what are the %d best sentences that should be used to perform the Google search? ` +
		`Respond only the a sentences list, NOTHING else! This list NEEDS to be a \n separated list Ex: sentence1 \n sentence2 \n sentence3... !` +
		`Do not enumerate the list. When possible, use a different language to each sentence. ` +
		`person/company context:"%s". research:"%s". languages:"%s"`
	sentenceAmount = 5
)

type sentenceService struct {
	queueGemini            interfaces.Queue
	queueGoogleSearch      interfaces.Queue
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l sentenceService) validateGeminiResponse(response string) ([]string, *string) {
	split := strings.Split(strings.ToLower(response), "\n")
	if len(split) == 0 {
		message := fmt.Sprintf("sentences with wrong format '%s'", response)
		return nil, &message
	}
	return split, nil
}
func (l sentenceService) validateOrchestratorData(request nosqlmodels.Request) error {
	var messages []string
	if request.Context == nil {
		messages = append(messages, `"context" is required in mongoDB to perform sentence service`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required in mongoDB to perform sentence service`)
	}
	if request.Locations == nil {
		messages = append(messages, `"locations" is required in mongoDB to perform sentence service`)
	}
	if request.Languages == nil {
		messages = append(messages, `"languages" is required in mongoDB to perform sentence service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l sentenceService) buildQuestion(request nosqlmodels.Request) string {
	return fmt.Sprintf(
		sentenceQuestionTemplate,
		sentenceAmount,
		*request.Context,
		*request.Research,
		strings.Join(*request.Languages, ","),
	)
}

func (l sentenceService) Execute(ctx context.Context, orchestratorRequest models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "sentenceService.Execute",
		slog.String("details", "process started"))

	var request nosqlmodels.Request
	err := l.orchestratorRepository.GetById(ctx, *orchestratorRequest.RequestId, &request)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateOrchestratorData(request)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question := l.buildQuestion(request)

	b, err := builder.BuildQueueGeminiMessage(
		*orchestratorRequest.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.Sentences,
		0,
	)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l sentenceService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "sentenceService.Callback",
		slog.String("details", "process started"))

	sentences, errMessage := l.validateGeminiResponse(*callback.Response)
	if errMessage != nil {
		var request nosqlmodels.Request
		err := l.orchestratorRepository.GetById(ctx, *callback.RequestId, &request)
		if err != nil {
			slog.ErrorContext(ctx, "sentenceService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
		question := l.buildQuestion(request)
		err = errortypes.NewInvalidAIResponseException(*callback.RequestId, question, enumactions.Sentences, callback.ReceiveCount+1, *errMessage)
		slog.ErrorContext(ctx, "sentenceService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err := l.orchestratorRepository.Update(ctx, *callback.RequestId,
		bson.M{"sentences": sentences},
	)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var b []byte
	b, err = builder.BuildQueueGoogleSearchMessage(*callback.RequestId)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGoogleSearch.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "sentenceService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	slog.InfoContext(ctx, "sentenceService.Callback",
		slog.String("details", "process finished"))
	return nil
}
func NewSentenceService(queueGemini, queueGoogleSearch interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository) interfaces.Service {
	return &sentenceService{
		queueGemini:            queueGemini,
		orchestratorRepository: orchestratorRepository,
		queueGoogleSearch:      queueGoogleSearch,
	}
}
