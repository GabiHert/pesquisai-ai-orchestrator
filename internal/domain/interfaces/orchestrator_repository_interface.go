package interfaces

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
)

type OrchestratorRepository interface {
	GetById(ctx context.Context, id string, model interface{}) error
	Create(ctx context.Context, model interface{}) error
	Update(ctx context.Context, id string, values bson.M) error
	Connect(database, collection string)
}
