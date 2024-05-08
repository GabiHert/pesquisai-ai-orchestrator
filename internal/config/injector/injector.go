package injector

import (
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/properties"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/delivery/controllers"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/factory"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/interfaces"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/services"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/domain/usecases"
	nosql "github.com/PesquisAi/pesquisai-database-lib/nosql/connection"
	nosqlrepositories "github.com/PesquisAi/pesquisai-database-lib/nosql/repositories"
	sql "github.com/PesquisAi/pesquisai-database-lib/sql/connection"
	sqlrepositories "github.com/PesquisAi/pesquisai-database-lib/sql/repositories"
	"github.com/PesquisAi/pesquisai-rabbitmq-lib/rabbitmq"
	"gorm.io/gorm"
	"net/http"
)

type Dependencies struct {
	Mux                                 *http.ServeMux
	Controller                          interfaces.Controller
	RequestRepository                   interfaces.RequestRepository
	OrchestratorRepository              interfaces.OrchestratorRepository
	DatabaseSqlConnection               *sql.Connection
	DatabaseNoSqlConnection             *nosql.Connection
	QueueConnection                     *rabbitmq.Connection
	UseCase                             interfaces.UseCase
	QueueGemini                         interfaces.Queue
	QueueStatusManager                  interfaces.Queue
	ConsumerAiOrchestratorQueue         interfaces.QueueConsumer
	QueueAiOrchestrator                 interfaces.Queue
	ConsumerAiOrchestratorCallbackQueue interfaces.QueueConsumer
	ServiceFactory                      *factory.ServiceFactory
}

func (d *Dependencies) Inject() *Dependencies {
	if d.DatabaseSqlConnection == nil {
		d.DatabaseSqlConnection = &sql.Connection{DB: &gorm.DB{}}
	}

	if d.DatabaseNoSqlConnection == nil {
		d.DatabaseNoSqlConnection = &nosql.Connection{}
	}

	if d.OrchestratorRepository == nil {
		d.OrchestratorRepository = &nosqlrepositories.Repository{}
	}

	if d.Mux == nil {
		d.Mux = http.NewServeMux()
	}

	if d.RequestRepository == nil {
		d.RequestRepository = &sqlrepositories.Request{Connection: d.DatabaseSqlConnection}
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

	if d.ConsumerAiOrchestratorQueue == nil || d.QueueAiOrchestrator == nil {
		queue := rabbitmq.NewQueue(
			d.QueueConnection,
			properties.QueueNameAiOrchestrator,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
		d.ConsumerAiOrchestratorQueue = queue
		d.QueueAiOrchestrator = queue
	}

	if d.ConsumerAiOrchestratorCallbackQueue == nil {
		d.ConsumerAiOrchestratorCallbackQueue = rabbitmq.NewQueue(
			d.QueueConnection,
			properties.QueueNameAiOrchestratorCallback,
			rabbitmq.CONTENT_TYPE_JSON,
			properties.CreateQueueIfNX())
	}

	if d.Controller == nil {
		d.Controller = controllers.NewController(d.UseCase)
	}
	return d
}

func NewDependencies() *Dependencies {
	deps := &Dependencies{}
	deps.Inject()
	return deps
}
