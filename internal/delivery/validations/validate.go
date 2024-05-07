package validations

import (
	"fmt"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/errortypes"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/dtos"
	"github.com/go-playground/validator/v10"
)

func getError(tag string, field string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "max":
		switch field {
		case "context":
			return "'context' should have at most '1000' characters"
		case "research":
			return "'research' should have at most '1000' characters"
		}
	case "min":
		switch field {
		case "context":
			return "'context' should have at least '100' characters"
		case "research":
			return "'research' should have at least '10' characters"
		}

	}

	return ""
}
func getField(field string) string {
	switch field {
	case "Research":
		return "research"
	case "Context":
		return "context"
	}
	return ""
}

func Validate(request *dtos.AiOrchestratorRequest) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(request)
	if err == nil {
		return nil
	}

	var messages []string
	for _, err := range err.(validator.ValidationErrors) {
		messages = append(messages, getError(err.ActualTag(), getField(err.Field())))
	}

	return errortypes.NewValidationException(messages...)
}
