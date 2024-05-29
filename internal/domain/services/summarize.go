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
	enumstatus "github.com/PesquisAi/pesquisai-database-lib/sql/enums/status"
	"log/slog"
)

const (
	summarizeQuestionTemplate = "You are a part of a major project that performs researches for business and you have one responsibility." +
		" To summarize the content of a webpage given the research purpose . You will receive a context about the " +
		"researcher and the research. Answer only with the summary and nothing else. Make the summary relatively short.\n" +
		"researcher context:%s\n" +
		"research:%s\n" +
		"Web page content:%s"
)

type summarizeService struct {
	queueStatusManager     interfaces.Queue
	queueGemini            interfaces.Queue
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l summarizeService) validateOrchestratorData(request nosqlmodels.Request) error {
	var messages []string
	if request.Context == nil {
		messages = append(messages, `"context" is required in mongoDB to perform sentence service`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required in mongoDB to perform sentence service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}
func (l summarizeService) validaCallbackRequest(request models.AiOrchestratorCallbackRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform summarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l summarizeService) validateOrchestratorRequest(request models.AiOrchestratorRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform summarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l summarizeService) validateResearch(research nosqlmodels.Research) error {
	var messages []string
	if research.Content == nil {
		messages = append(messages, `"content" is required in research table to perform summarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l summarizeService) buildQuestion(ctx context.Context, requestId, researchId string) (question string, err error) {

	var research nosqlmodels.Research
	err = l.orchestratorRepository.GetById(ctx, researchId, &research)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	err = l.validateResearch(research)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	var request nosqlmodels.Request
	err = l.orchestratorRepository.GetById(ctx, requestId, &request)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	err = l.validateOrchestratorData(request)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	return fmt.Sprintf(
		summarizeQuestionTemplate,
		*request.Context,
		*request.Research,
		*research.Content,
	), nil
}

func (l summarizeService) Execute(ctx context.Context, orchestratorRequest models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "summarizeService.Execute",
		slog.String("details", "process started"))

	err := l.validateOrchestratorRequest(orchestratorRequest)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question, err := l.buildQuestion(ctx, *orchestratorRequest.RequestId, *orchestratorRequest.ResearchId)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	b, err := builder.BuildQueueGeminiMessage(
		*orchestratorRequest.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.Summarize,
		0,
	)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l summarizeService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "summarizeService.Callback",
		slog.String("details", "process started"))

	err := l.validaCallbackRequest(callback)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.orchestratorRepository.Update(ctx, *callback.ResearchId, map[string]any{
		"summary": *callback.Response,
	})
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var b []byte
	b, err = builder.BuildQueueStatusManagerMessage(nil, callback.ResearchId, enumstatus.SUMMARIZED)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueStatusManager.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "summarizeService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	slog.InfoContext(ctx, "summarizeService.Callback",
		slog.String("details", "process finished"))
	return nil
}
func NewSummarizeService(queueGemini, queueStatusManager interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository) interfaces.Service {
	return &summarizeService{
		queueStatusManager:     queueStatusManager,
		queueGemini:            queueGemini,
		orchestratorRepository: orchestratorRepository,
	}
}
