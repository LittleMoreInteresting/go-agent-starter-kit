package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-agent-starter-kit/internal/agent"
	"go-agent-starter-kit/internal/config"
	"go-agent-starter-kit/internal/knowledge"
	"go-agent-starter-kit/internal/llm"
	"go-agent-starter-kit/internal/memory"
	"go-agent-starter-kit/internal/server"
	"go-agent-starter-kit/internal/tools"
	kitlog "go-agent-starter-kit/pkg/logger"
)

func main() {
	logger := kitlog.New()
	cfgPath := envOr("CONFIG_PATH", "configs/config.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	var mem memory.Store
	switch cfg.Memory.Backend {
	case "redis":
		mem = memory.NewRedisStore(cfg.Memory.RedisAddr, cfg.Memory.RedisPass, cfg.Memory.KeyPrefix, cfg.Memory.RedisDB)
	default:
		mem = memory.NewInMemoryStore()
	}
	defer mem.Close()

	docs, err := knowledge.LoadDocuments(cfg.Agent.KnowledgePath)
	if err != nil {
		logger.Printf("knowledge base unavailable (%v), fallback to empty", err)
	}
	kb := knowledge.NewInMemoryBase(docs)

	llmClient := llm.NewOpenAIClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, time.Duration(cfg.LLM.TimeoutS)*time.Second)
	registry := tools.NewRegistry(tools.NewCalculatorTool())
	exec := agent.NewExecutor(registry)
	ag := agent.New(llmClient, mem, exec, cfg.Memory.HistorySize, cfg.Agent.SystemPrompt, kb, cfg.Agent.KnowledgeTopK)

	h := server.New(ag).Handler()
	srv := &http.Server{Addr: cfg.Server.Address, Handler: h}

	go func() {
		logger.Printf("server listening on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()

	waitForShutdown(logger, srv)
}

func waitForShutdown(logger *log.Logger, srv *http.Server) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
	logger.Println("graceful shutdown complete")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
