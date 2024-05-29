package interfaces

import (
	"context"
	"github.com/PesquisAi/pesquisai-database-lib/sql/models"
)

type ResearchRepository interface {
	Get(ctx context.Context, id string) (research *models.Research, err error)
}
