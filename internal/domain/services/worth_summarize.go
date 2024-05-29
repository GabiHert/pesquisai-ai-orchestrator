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
	enumstatus "github.com/PesquisAi/pesquisai-database-lib/sql/enums/status"
	"log/slog"
	"strings"
)

const (
	worthSummarizeQuestionTemplate = "You are a part of a major project that performs researches for business and you have one responsibility. " +
		"To determine if the content in the webpage is essential for the researcher." +
		" It really needs to be essential for you to consider it." +
		" If its not, you should answer me with the letter N and nothing else. " +
		"If the content is really important given the research, answer me with a Y and nothing else." +
		" You will receive a context about the researcher and the research.\n" +
		"researcher context:%s\n" +
		"research:%s\n" +
		"Web page content:%s"
)

type worthSummarizeService struct {
	queueStatusManager     interfaces.Queue
	queueOrchestrator      interfaces.Queue
	queueGemini            interfaces.Queue
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l worthSummarizeService) validateGeminiResponse(response string) (bool, *string) {
	res := strings.ToLower(response)
	if res != "n" && res != "y" {
		message := fmt.Sprintf("worth acessing response with wrong format '%s'", response)
		return false, &message
	}
	return res == "y", nil
}

func (l worthSummarizeService) validateOrchestratorData(request nosqlmodels.Request) error {
	var messages []string
	if request.Context == nil {
		messages = append(messages, `"context" is required in mongoDB to perform sentence service`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required in mongoDB to perform sentence service`)
	}
	if request.Research == nil {
		messages = append(messages, `"research" is required in mongoDB to perform sentence service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}
func (l worthSummarizeService) validaCallbackRequest(request models.AiOrchestratorCallbackRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform worthSummarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthSummarizeService) validateOrchestratorRequest(request models.AiOrchestratorRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform worthSummarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthSummarizeService) validateResearch(research nosqlmodels.Research) error {
	var messages []string
	if research.Content == nil {
		messages = append(messages, `"content" is required in research table to perform worthSummarize service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthSummarizeService) buildQuestion(ctx context.Context, requestId string, research nosqlmodels.Research) (question string, err error) {

	var request nosqlmodels.Request
	err = l.orchestratorRepository.GetById(ctx, requestId, &request)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	err = l.validateOrchestratorData(request)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	return fmt.Sprintf(
		worthSummarizeQuestionTemplate,
		*request.Context,
		*request.Research,
		*research.Content,
	), nil
}

func (l worthSummarizeService) Execute(ctx context.Context, orchestratorRequest models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "worthSummarizeService.Execute",
		slog.String("details", "process started"))

	err := l.validateOrchestratorRequest(orchestratorRequest)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var research nosqlmodels.Research
	err = l.orchestratorRepository.GetById(ctx, *orchestratorRequest.ResearchId, &research)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateResearch(research)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question, err := l.buildQuestion(ctx, *orchestratorRequest.RequestId, research)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	b, err := builder.BuildQueueGeminiMessage(
		*orchestratorRequest.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.WorthSummarize,
		0,
	)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l worthSummarizeService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "worthSummarizeService.Callback",
		slog.String("details", "process started"))

	err := l.validaCallbackRequest(callback)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var research nosqlmodels.Research
	err = l.orchestratorRepository.GetById(ctx, *callback.ResearchId, &research)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateResearch(research)
	if err != nil {
		slog.ErrorContext(ctx, "worthSummarizeService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	worth, errMessage := l.validateGeminiResponse(*callback.Response)
	if errMessage != nil {
		var request nosqlmodels.Request
		err = l.orchestratorRepository.GetById(ctx, *callback.RequestId, &request)
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		var question string
		question, err = l.buildQuestion(ctx, *callback.RequestId, research)
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Execute",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = errortypes.NewInvalidAIResponseException(*callback.RequestId, question, enumactions.WorthSummarize, callback.ReceiveCount+1, *errMessage)
		slog.ErrorContext(ctx, "worthSummarizeService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var b []byte
	if worth {
		action := enumactions.Summarize
		b, err = builder.BuildQueueOrchestratorMessage(dtos.AiOrchestratorRequest{
			RequestId:  callback.RequestId,
			ResearchId: callback.ResearchId,
			Action:     &action,
		})
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = l.queueOrchestrator.Publish(ctx, b)
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
	} else {
		b, err = builder.BuildQueueStatusManagerMessage(nil, callback.ResearchId, enumstatus.FINISHED)
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = l.queueStatusManager.Publish(ctx, b)
		if err != nil {
			slog.ErrorContext(ctx, "worthSummarizeService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
	}

	slog.InfoContext(ctx, "worthSummarizeService.Callback",
		slog.String("details", "process finished"))
	return nil
}
func NewWorthSummarizeService(queueGemini, queueOrchestrator, queueStatusManager interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository) interfaces.Service {
	return &worthSummarizeService{
		queueStatusManager:     queueStatusManager,
		queueOrchestrator:      queueOrchestrator,
		queueGemini:            queueGemini,
		orchestratorRepository: orchestratorRepository,
	}
}
