package factory

import (
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	enumactions "github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/enums/actions"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
)

type ServiceFactory struct {
	LocationService       interfaces.Service
	LanguageService       interfaces.Service
	SentencesService      interfaces.Service
	WorthAccessingService interfaces.Service
	WorthSummarizeService interfaces.Service
	SummarizeService      interfaces.Service
}

func (sf ServiceFactory) Factory(action string) (interfaces.Service, error) {
	switch action {
	case enumactions.Location:
		return sf.LocationService, nil
	case enumactions.Language:
		return sf.LanguageService, nil
	case enumactions.Sentences:
		return sf.SentencesService, nil
	case enumactions.WorthAccessing:
		return sf.WorthAccessingService, nil
	case enumactions.WorthSummarize:
		return sf.WorthSummarizeService, nil
	case enumactions.Summarize:
		return sf.SummarizeService, nil
	}
	return nil, errortypes.NewServiceNotFoundException(fmt.Sprintf("Service for action '%s' not found", action))
}
