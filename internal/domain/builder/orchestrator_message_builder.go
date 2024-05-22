package builder

import (
	"encoding/json"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/dtos"
)

func BuildQueueOrchestratorMessage(orchestratorDto dtos.AiOrchestratorRequest) ([]byte, error) {
	return json.Marshal(&orchestratorDto)
}
