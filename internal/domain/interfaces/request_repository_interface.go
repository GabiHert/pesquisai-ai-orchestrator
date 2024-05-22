package interfaces

import (
	"context"
	"github.com/PesquisAi/pesquisai-database-lib/sql/models"
)

type RequestRepository interface {
	Create(ctx context.Context, request *models.Request) error
	GetWithRelations(ctx context.Context, id string) (request *models.Request, err error)
	RelateLanguage(ctx context.Context, id string, language string) error
	RelateLocation(ctx context.Context, id string, location string) error
}
