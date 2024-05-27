package main

import (
	"context"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/connections"
	"github.com/PesquisAi/pesquisai-ai-orchestrator/internal/config/injector"
	"github.com/joho/godotenv"
	"log/slog"
	"sync"
)

func main() {
	var err error

	if err = godotenv.Load(".env"); err != nil {
		panic(err)
	}
	deps := injector.NewDependencies()

	if err = connections.Connect(deps); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if err := deps.ConsumerAiOrchestratorQueue.Consume(context.Background(), deps.Controller.AiOrchestratorHandler); err != nil {
			slog.Error("Error during ai-orchestrator-callback routine: ", err)
			wg.Done()
		}
	}()

	go func() {
		if err := deps.ConsumerAiOrchestratorCallbackQueue.Consume(context.TODO(), deps.Controller.AiOrchestratorCallbackHandler); err != nil {
			slog.Error("Error during ai-orchestrator-callback routine: ", err)
			wg.Done()
		}
	}()

	wg.Wait()

	if err = connections.Disconnect(deps); err != nil {
		panic(err)
	}
}
