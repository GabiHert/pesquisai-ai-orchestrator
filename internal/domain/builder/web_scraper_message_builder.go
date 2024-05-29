package builder

import (
	"encoding/json"
)

type queueMessage struct {
	RequestId  string `json:"request_id"`
	ResearchId string `json:"research_id"`
	Url        string `json:"url"`
}

func BuildQueueWebScraperMessage(requestId, researchId, url string) ([]byte, error) {
	msg := &queueMessage{
		RequestId:  requestId,
		ResearchId: researchId,
		Url:        url,
	}

	return json.Marshal(msg)
}
