package builder

import (
	"encoding/json"
)

type message struct {
	RequestId   *string         `json:"request_id"`
	Question    *string         `json:"question"`
	OutputQueue *string         `json:"output_queue"`
	Forward     *map[string]any `json:"forward"`
}

func BuildQueueGeminiMessage(requestId, question, outputQueue, action string, receiveCount int) ([]byte, error) {

	msg := &message{
		RequestId:   &requestId,
		Question:    &question,
		OutputQueue: &outputQueue,
		Forward: &map[string]any{
			"action":        action,
			"receive_count": receiveCount,
		},
	}

	return json.Marshal(msg)
}
