package properties

import "os"

const (
	DatabaseTablePrefix             = "pesquisai."
	QueueNameGemini                 = "gemini"
	QueueNameAiOrchestrator         = "ai-orchestrator"
	QueueNameAiOrchestratorCallback = "ai-orchestrator-callback"
	QueueNameStatusManager          = "status-manager"
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

func DatabaseConnectionUser() string {
	return os.Getenv("DATABASE_CONNECTION_USER")
}

func DatabaseConnectionHost() string {
	return os.Getenv("DATABASE_CONNECTION_HOST")
}

func DatabaseConnectionName() string {
	return os.Getenv("DATABASE_CONNECTION_NAME")
}

func DatabaseConnectionPort() string {
	return os.Getenv("DATABASE_CONNECTION_PORT")
}

func DatabaseConnectionPassword() string {
	return os.Getenv("DATABASE_CONNECTION_PASSWORD")
}
