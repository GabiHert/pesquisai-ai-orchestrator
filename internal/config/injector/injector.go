package injector

import (
	"github.com/PesquisAi/pesquisai-api/internal/config/properties"
	"github.com/PesquisAi/pesquisai-api/internal/domain/factory"
	"github.com/PesquisAi/pesquisai-api/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-api/internal/domain/services"
	"github.com/PesquisAi/pesquisai-api/internal/domain/usecases"
	"github.com/PesquisAi/pesquisai-database-lib/connection"
	"github.com/PesquisAi/pesquisai-database-lib/repositories"
	"github.com/PesquisAi/pesquisai-rabbitmq-lib/rabbitmq"
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
	ServiceFactory                      *factory.ServiceFactory
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

	if d.ServiceFactory == nil {
		d.ServiceFactory = &factory.ServiceFactory{
			LocationService:         services.NewLocationService(d.QueueGemini),
			LanguageService:         nil,
			SentencesService:        nil,
			WorthCheckingService:    nil,
			WorthSummarizingService: nil,
			SummarizeService:        nil,
		}
	}

	if d.UseCase == nil {
		d.UseCase = usecases.NewUseCase(d.RequestRepository, d.ServiceFactory)
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
