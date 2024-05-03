package dtos

type AiOrchestratorRequest struct {
	RequestId *string `json:"request_id" validate:"uuid,required"`
	Context   *string `json:"context" validate:"required"`
	Research  *string `json:"research" validate:"required"`
	Action    *string `json:"action" validate:"required,oneof= location language sentences worth-checking worth-summarizing summarize"`
}
