package errortypes

import (
	"github.com/GabiHert/pesquisai-errors-lib/exceptions"
	"net/http"
)

func NewUnknownException(message string) *exceptions.Error {
	return &exceptions.Error{Messages: []string{message}, ErrorType: exceptions.ErrorType{
		Code:           "PAPI01",
		Type:           "Unknown",
		HttpStatusCode: http.StatusInternalServerError,
	}}
}

func NewValidationException(messages ...string) *exceptions.Error {
	return &exceptions.Error{
		Messages: messages,
		ErrorType: exceptions.ErrorType{
			Code:           "PAPI02",
			Type:           "Validation",
			HttpStatusCode: http.StatusBadRequest,
		}}
}

func NewNotFoundException(messages ...string) *exceptions.Error {
	return &exceptions.Error{
		Messages: messages,
		ErrorType: exceptions.ErrorType{
			Code:           "PAPI03",
			Type:           "Not found",
			HttpStatusCode: http.StatusNotFound,
		}}
}
