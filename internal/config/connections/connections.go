package connections

import (
	"github.com/PesquisAi/pesquisai-api/internal/config/injector"
	"github.com/PesquisAi/pesquisai-api/internal/config/properties"
	"github.com/PesquisAi/pesquisai-database-lib/connection"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Connect(deps *injector.Dependencies) error {
	err := deps.DatabaseConnection.Connect(connection.Config{
		User: properties.DatabaseConnectionUser(),
		Host: properties.DatabaseConnectionHost(),
		Psw:  properties.DatabaseConnectionPassword(),
		Name: properties.DatabaseConnectionName(),
		Port: properties.DatabaseConnectionPort(),
		GormConfig: gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: properties.DatabaseTablePrefix,
			},
		},
	})
	if err != nil {
		return err
	}

	err = deps.QueueConnection.Connect(
		properties.QueueConnectionUser(),
		properties.QueueConnectionPassword(),
		properties.QueueConnectionHost(),
		properties.QueueConnectionPort())
	if err != nil {
		return err
	}

	err = deps.QueueGemini.Connect(properties.QueueNameGemini)
	if err != nil {
		return err
	}

	err = deps.ConsumerAiOrchestratorQueue.Connect(properties.QueueNameAiOrchestrator)
	if err != nil {
		return err
	}

	return nil
}
