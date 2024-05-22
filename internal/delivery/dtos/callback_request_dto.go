package dtos

type AiOrchestratorCallbackRequest struct {
	RequestId  *string `json:"request_id,omitempty" validate:"uuid"`
	ResearchId *string `json:"research_id,omitempty" validate:"omitempty,uuid"`
	Response   *string `json:"response,omitempty" validate:"required"`
	Forward    *struct {
		Action *string `json:"action" validate:"required,oneof= location language sentences worth-checking worth-summarizing summarize"`
	} `json:"forward" validate:"required"`
}
