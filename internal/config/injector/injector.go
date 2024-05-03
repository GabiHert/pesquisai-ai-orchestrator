package injector

import (
	"github.com/GabiHert/pesquisai-api/internal/config/properties"
	"github.com/GabiHert/pesquisai-api/internal/domain/interfaces"
	"github.com/GabiHert/pesquisai-api/internal/domain/usecases"
	"github.com/GabiHert/pesquisai-database-lib/connection"
	"github.com/GabiHert/pesquisai-database-lib/repositories"
	"github.com/GabiHert/pesquisai-rabbitmq-lib/rabbitmq"
	"gorm.io/gorm"
	"net/http"
)

type Dependencies struct {
	Mux                                 *http.ServeMux
	Controller                          interfaces.Controller
	RequestRepository                   interfaces.RequestRepository
	DatabaseConnection                  *connection.Connection
	QueueConnection                     *rabbitmq.Connection
	UseCase                             interfaces.UseCase
	QueueGemini                         interfaces.Queue
	QueueStatusManager                  interfaces.Queue
	ConsumerAiOrchestratorQueue         interfaces.QueueConsumer
	ConsumerAiOrchestratorCallbackQueue interfaces.QueueConsumer
}

func (d *Dependencies) Inject() *Dependencies {
	if d.DatabaseConnection == nil {
		d.DatabaseConnection = &connection.Connection{DB: &gorm.DB{}}
	}

	if d.Mux == nil {
		d.Mux = http.NewServeMux()
	}

	if d.RequestRepository == nil {
		d.RequestRepository = &repositories.Request{Connection: d.DatabaseConnection}
	}

	if d.QueueConnection == nil {
		d.QueueConnection = &rabbitmq.Connection{}
	}

	if d.QueueGemini == nil {
		d.QueueGemini = rabbitmq.NewQueue(d.QueueConnection,
			properties.QueueNameGemini,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
	}

	if d.QueueStatusManager == nil {
		d.QueueStatusManager = rabbitmq.NewQueue(d.QueueConnection,
			properties.QueueNameStatusManager,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
	}

	if d.UseCase == nil {
		d.UseCase = usecases.NewUseCase(d.RequestRepository, d.QueueGemini)
	}

	if d.ConsumerAiOrchestratorQueue == nil {
		d.ConsumerAiOrchestratorQueue = rabbitmq.NewQueue(
			d.QueueConnection,
			properties.QueueNameAiOrchestrator,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
	}

	if d.ConsumerAiOrchestratorCallbackQueue == nil {
		d.ConsumerAiOrchestratorCallbackQueue = rabbitmq.NewQueue(
			d.QueueConnection,
			properties.QueueNameAiOrchestratorCallback,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
	}

	if d.Controller == nil {
		//	d.Controller = controllers.NewController(d.UseCase)
	}
	return d
}

func NewDependencies() *Dependencies {
	deps := &Dependencies{}
	deps.Inject()
	return deps
}
