package dtos

type AiOrchestratorRequest struct {
	RequestId  *string `json:"request_id" validate:"uuid,required"`
	ResearchId *string `json:"research_id"`
	Context    *string `json:"context"`
	Research   *string `json:"research"`
	Action     *string `json:"action" validate:"required,oneof= location language sentences worth-checking worth-summarizing summarize"`
}
