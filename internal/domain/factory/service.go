package factory

import (
	enumactions "github.com/GabiHert/pesquisai-api/internal/domain/enums/actions"
	"github.com/GabiHert/pesquisai-api/internal/domain/interfaces"
	"github.com/GabiHert/pesquisai-api/internal/domain/models"
)

func FactorService(request models.AiOrchestratorRequest) (interfaces.Service, error) {
	switch *request.Action {
	case enumactions.LOCATION:
	case enumactions.LANGUAGE:
	case enumactions.SENTENCES:
	case enumactions.WORTH_CHECKING:
	case enumactions.WORTH_SUMMARIZING:
	case enumactions.SUMMARIZE:
	}

	return nil, nil
}
