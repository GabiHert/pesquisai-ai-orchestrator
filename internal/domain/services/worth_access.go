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
	"strings"
)

const (
	worthAccessQuestionTemplate = "You are part of a major project that performs researches for business and your only responsibility is to say" +
		" Y for Yes and N for No if a web page is worth accessing given the context about the researcher, the research and the web page title and url.\n" +
		"researcher context:%s.\n" +
		"research:%s\n" +
		"title:%s\n" +
		"url:%s"
)

type worthAccessingService struct {
	queueStatusManager     interfaces.Queue
	queueWebScraper        interfaces.Queue
	queueGemini            interfaces.Queue
	orchestratorRepository interfaces.OrchestratorRepository
}

func (l worthAccessingService) validateGeminiResponse(response string) (bool, *string) {
	res := strings.ToLower(response)
	if res != "n" && res != "y" {
		message := fmt.Sprintf("worth acessing response with wrong format '%s'", response)
		return false, &message
	}
	return res == "y", nil
}

func (l worthAccessingService) validateOrchestratorData(request nosqlmodels.Request) error {
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
func (l worthAccessingService) validaCallbackRequest(request models.AiOrchestratorCallbackRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform worthAccessing service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthAccessingService) validateOrchestratorRequest(request models.AiOrchestratorRequest) error {
	var messages []string
	if request.ResearchId == nil {
		messages = append(messages, `"research_id" is required to perform worthAccessing service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthAccessingService) validateResearch(request nosqlmodels.Research) error {
	var messages []string
	if request.Link == nil {
		messages = append(messages, `"link" is required in research table to perform worthAccessing service`)
	}
	if request.Title == nil {
		messages = append(messages, `"title" is required in research table to perform worthAccessing service`)
	}
	if len(messages) > 0 {
		return errortypes.NewValidationException(messages...)
	}
	return nil
}

func (l worthAccessingService) buildQuestion(ctx context.Context, requestId string, research nosqlmodels.Research) (question string, err error) {

	var request nosqlmodels.Request
	err = l.orchestratorRepository.GetById(ctx, requestId, &request)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	err = l.validateOrchestratorData(request)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return
	}

	return fmt.Sprintf(
		worthAccessQuestionTemplate,
		*request.Context,
		*request.Research,
		*research.Title,
		*research.Link,
	), nil
}

func (l worthAccessingService) Execute(ctx context.Context, orchestratorRequest models.AiOrchestratorRequest) error {
	slog.InfoContext(ctx, "worthAccessingService.Execute",
		slog.String("details", "process started"))

	err := l.validateOrchestratorRequest(orchestratorRequest)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var research nosqlmodels.Research
	err = l.orchestratorRepository.GetById(ctx, *orchestratorRequest.ResearchId, &research)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateResearch(research)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	question, err := l.buildQuestion(ctx, *orchestratorRequest.RequestId, research)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	b, err := builder.BuildQueueGeminiMessage(
		*orchestratorRequest.RequestId,
		question,
		properties.QueueNameAiOrchestratorCallback,
		enumactions.WorthAccessing,
		0,
	)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.queueGemini.Publish(ctx, b)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (l worthAccessingService) Callback(ctx context.Context, callback models.AiOrchestratorCallbackRequest) error {
	slog.InfoContext(ctx, "worthAccessingService.Callback",
		slog.String("details", "process started"))

	err := l.validaCallbackRequest(callback)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var research nosqlmodels.Research
	err = l.orchestratorRepository.GetById(ctx, *callback.ResearchId, &research)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	err = l.validateResearch(research)
	if err != nil {
		slog.ErrorContext(ctx, "worthAccessingService.Execute",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	worth, errMessage := l.validateGeminiResponse(*callback.Response)
	if errMessage != nil {
		var request nosqlmodels.Request
		err = l.orchestratorRepository.GetById(ctx, *callback.RequestId, &request)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		var question string
		question, err = l.buildQuestion(ctx, *callback.RequestId, research)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Execute",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = errortypes.NewInvalidAIResponseException(*callback.RequestId, question, enumactions.WorthAccessing, callback.ReceiveCount+1, *errMessage)
		slog.ErrorContext(ctx, "worthAccessingService.Callback",
			slog.String("details", "process error"),
			slog.String("error", err.Error()))
		return err
	}

	var b []byte
	if worth {
		b, err = builder.BuildQueueWebScraperMessage(*callback.RequestId, *callback.ResearchId, *research.Link)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = l.queueWebScraper.Publish(ctx, b)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
	} else {
		b, err = builder.BuildQueueStatusManagerMessage(nil, callback.ResearchId, enumstatus.FINISHED)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}

		err = l.queueStatusManager.Publish(ctx, b)
		if err != nil {
			slog.ErrorContext(ctx, "worthAccessingService.Callback",
				slog.String("details", "process error"),
				slog.String("error", err.Error()))
			return err
		}
	}

	slog.InfoContext(ctx, "worthAccessingService.Callback",
		slog.String("details", "process finished"))
	return nil
}
func NewWorthAccessingService(queueGemini, queueWebScraper, queueStatusManager interfaces.Queue, orchestratorRepository interfaces.OrchestratorRepository) interfaces.Service {
	return &worthAccessingService{
		queueStatusManager:     queueStatusManager,
		queueWebScraper:        queueWebScraper,
		queueGemini:            queueGemini,
		orchestratorRepository: orchestratorRepository,
	}
}
