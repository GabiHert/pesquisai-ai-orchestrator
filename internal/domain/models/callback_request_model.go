package models

type AiOrchestratorCallbackRequest struct {
	RequestId    *string
	ResearchId   *string
	Response     *string
	Action       *string
	ReceiveCount int
}
