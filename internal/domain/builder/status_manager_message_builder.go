package builder

import (
	"encoding/json"
)

type statusManagerMessage struct {
	RequestId  *string `json:"request_id,omitempty"`
	ResearchId *string `json:"research_id,omitempty"`
	Status     string  `json:"status"`
}

func BuildQueueStatusManagerMessage(requestId, researchId *string, status string) ([]byte, error) {
	msg := &statusManagerMessage{
		RequestId:  requestId,
		ResearchId: researchId,
		Status:     status,
	}

	return json.Marshal(msg)
}
