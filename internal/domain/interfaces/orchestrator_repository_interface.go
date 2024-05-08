package interfaces

import (
	"context"
)

type OrchestratorRepository interface {
	GetById(ctx context.Context, id string, model interface{}) error
	Create(ctx context.Context, model interface{}) error
}
