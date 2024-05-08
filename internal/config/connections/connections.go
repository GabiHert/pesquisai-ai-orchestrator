package connections

import (
	"context"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/injector"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-database-lib/sql/connection"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Connect(deps *injector.Dependencies) error {
	err := deps.DatabaseSqlConnection.Connect(connection.Config{
		User: properties.DatabaseSqlConnectionUser(),
		Host: properties.DatabaseSqlConnectionHost(),
		Psw:  properties.DatabaseSqlConnectionPassword(),
		Name: properties.DatabaseSqlConnectionName(),
		Port: properties.DatabaseSqlConnectionPort(),
		GormConfig: gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: properties.DatabaseTablePrefix,
			},
		},
	})
	if err != nil {
		return err
	}

	err = deps.DatabaseNoSqlConnection.Connect(context.Background(),
		properties.DatabaseNoSqlConnectionHost(),
		properties.DatabaseNoSqlConnectionPort(),
	)
	if err != nil {
		return err
	}

	err = deps.QueueConnection.Connect(
		properties.QueueConnectionUser(),
		properties.QueueConnectionPassword(),
		properties.QueueConnectionHost(),
		properties.QueueConnectionPort(),
	)
	if err != nil {
		return err
	}

	err = deps.QueueGemini.Connect()
	if err != nil {
		return err
	}

	err = deps.ConsumerAiOrchestratorQueue.Connect()
	if err != nil {
		return err
	}

	err = deps.ConsumerAiOrchestratorCallbackQueue.Connect()
	if err != nil {
		return err
	}

	return nil
}
