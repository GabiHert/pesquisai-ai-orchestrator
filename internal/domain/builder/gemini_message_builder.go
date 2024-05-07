package builder

import (
	"encoding/json"
)

type message struct {
	RequestId   *string         `json:"request_id"`
	Question    *string         `json:"question"`
	OutputQueue *string         `json:"outputQueue"`
	Forward     *map[string]any `json:"forward"`
}

func BuildQueueGeminiMessage(requestId, question, outputQueue string) ([]byte, error) {

	msg := &message{
		RequestId:   &requestId,
		Question:    &question,
		OutputQueue: &outputQueue,
	}

	return json.Marshal(msg)
}
