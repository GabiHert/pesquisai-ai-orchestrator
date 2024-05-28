package properties

import "os"

const (
	DatabaseTablePrefix             = "pesquisai."
	QueueNameGemini                 = "gemini"
	QueueNameGoogleSearch           = "google-search"
	QueueNameAiOrchestrator         = "ai-orchestrator"
	QueueNameAiOrchestratorCallback = "ai-orchestrator-callback"
	QueueNameStatusManager          = "status-manager"

	DatabaseNoSqlName                  = "pesquisai"
	DatabaseOrchestratorCollectionName = "orchestrator"
)

func CreateQueueIfNX() bool {
	return os.Getenv("CREATE_QUEUE_IF_NX") == "true"
}

func QueueConnectionUser() string {
	return os.Getenv("QUEUE_CONNECTION_USER")
}

func QueueConnectionPort() string {
	return os.Getenv("QUEUE_CONNECTION_PORT")
}

func QueueConnectionHost() string {
	return os.Getenv("QUEUE_CONNECTION_HOST")
}

func QueueConnectionPassword() string {
	return os.Getenv("QUEUE_CONNECTION_PASSWORD")
}

func DatabaseSqlConnectionUser() string {
	return os.Getenv("DATABASE_SQL_CONNECTION_USER")
}

func DatabaseSqlConnectionHost() string {
	return os.Getenv("DATABASE_SQL_CONNECTION_HOST")
}

func DatabaseSqlConnectionName() string {
	return os.Getenv("DATABASE_SQL_CONNECTION_NAME")
}

func DatabaseSqlConnectionPort() string {
	return os.Getenv("DATABASE_SQL_CONNECTION_PORT")
}

func DatabaseSqlConnectionPassword() string {
	return os.Getenv("DATABASE_SQL_CONNECTION_PASSWORD")
}

func DatabaseNoSqlConnectionHost() string {
	return os.Getenv("DATABASE_NO_SQL_CONNECTION_HOST")
}
func DatabaseNoSqlConnectionPort() string {
	return os.Getenv("DATABASE_NO_SQL_CONNECTION_PORT")
}
