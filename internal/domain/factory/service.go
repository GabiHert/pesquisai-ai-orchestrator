package factory

import (
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	enumactions "github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/enums/actions"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/models"
)

type ServiceFactory struct {
	LocationService         interfaces.Service
	LanguageService         interfaces.Service
	SentencesService        interfaces.Service
	WorthCheckingService    interfaces.Service
	WorthSummarizingService interfaces.Service
	SummarizeService        interfaces.Service
}

func (sf ServiceFactory) Factory(request models.AiOrchestratorRequest) (interfaces.Service, error) {
	switch *request.Action {
	case enumactions.LOCATION:
		return sf.LocationService, nil
	case enumactions.LANGUAGE:
		return sf.LanguageService, nil
	case enumactions.SENTENCES:
		return sf.SentencesService, nil
	case enumactions.WORTH_CHECKING:
		return sf.WorthCheckingService, nil
	case enumactions.WORTH_SUMMARIZING:
		return sf.WorthSummarizingService, nil
	case enumactions.SUMMARIZE:
		return sf.SummarizeService, nil
	}
	return nil, errortypes.NewServiceNotFoundException(fmt.Sprintf("Service for action '%s' not found", *request.Action))
}
