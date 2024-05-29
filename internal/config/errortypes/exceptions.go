package errortypes

import (
	"github.com/PesquisAi/pesquisai-errors-lib/exceptions"
	"net/http"
)

const (
	UnknownCode           = "PAAO01"
	ValidateCode          = "PAAO02"
	ServiceNotFoundCode   = "PAAO03"
	InvalidAiResponseCode = "PAAO04"
)

func NewUnknownException(message string) *exceptions.Error {
	return &exceptions.Error{Messages: []string{message}, ErrorType: exceptions.ErrorType{
		Code:           UnknownCode,
		Type:           "Unknown",
		HttpStatusCode: http.StatusInternalServerError,
	}}
}

func NewValidationException(messages ...string) *exceptions.Error {
	return &exceptions.Error{
		Messages: messages,
		ErrorType: exceptions.ErrorType{
			Code:           ValidateCode,
			Type:           "Validation",
			HttpStatusCode: http.StatusBadRequest,
		}}
}

func NewServiceNotFoundException(messages ...string) *exceptions.Error {
	return &exceptions.Error{
		Messages: messages,
		ErrorType: exceptions.ErrorType{
			Code:           ServiceNotFoundCode,
			Type:           "Service could not be found to execute",
			HttpStatusCode: http.StatusNotFound,
		}}
}

func NewInvalidAIResponseException(requestId, question, action string, receiveCount int, messages ...string) *exceptions.Error {
	return &exceptions.Error{
		Messages: messages,
		Forward: map[string]any{
			"requestId":    requestId,
			"question":     question,
			"action":       action,
			"receiveCount": receiveCount,
		},
		ErrorType: exceptions.ErrorType{
			Code:           InvalidAiResponseCode,
			Type:           "Invalid AI response",
			HttpStatusCode: http.StatusBadRequest,
		},
	}
}
