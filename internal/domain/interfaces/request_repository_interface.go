package interfaces

import (
	"context"
	"github.com/GabiHert/pesquisai-database-lib/models"
)

type RequestRepository interface {
	Create(ctx context.Context, request *models.Request) error
	GetWithResearches(ctx context.Context, id string) (request *models.Request, err error)
}
