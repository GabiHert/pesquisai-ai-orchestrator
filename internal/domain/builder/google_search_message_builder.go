package builder

import (
	"encoding/json"
)

type googleSearchMessage struct {
	RequestId string `json:"request_id"`
}

func BuildQueueGoogleSearchMessage(requestId string) ([]byte, error) {
	return json.Marshal(&googleSearchMessage{RequestId: requestId})
}
