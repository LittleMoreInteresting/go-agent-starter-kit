package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server ServerConfig
	LLM    LLMConfig
	Memory MemoryConfig
	Agent  AgentConfig
}

type ServerConfig struct{ Address string }

type LLMConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
	TimeoutS int
}

type MemoryConfig struct {
	Backend     string
	RedisAddr   string
	RedisDB     int
	RedisPass   string
	KeyPrefix   string
	HistorySize int
}

type AgentConfig struct {
	SystemPrompt  string
	KnowledgePath string
	KnowledgeTopK int
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := parseSimpleYAML(string(b), cfg); err != nil {
		return nil, err
	}
	applyEnv(cfg)
	applyDefaults(cfg)
	return cfg, nil
}

func parseSimpleYAML(content string, cfg *Config) error {
	section := ""
	s := bufio.NewScanner(strings.NewReader(content))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, ":") {
			section = strings.TrimSuffix(line, ":")
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if idx := strings.Index(v, "#"); idx >= 0 {
			v = strings.TrimSpace(v[:idx])
		}
		v = strings.Trim(v, "\"")
		switch section {
		case "server":
			if k == "address" {
				cfg.Server.Address = v
			}
		case "llm":
			switch k {
			case "provider":
				cfg.LLM.Provider = v
			case "base_url":
				cfg.LLM.BaseURL = v
			case "api_key":
				cfg.LLM.APIKey = v
			case "model":
				cfg.LLM.Model = v
			case "timeout_seconds":
				if n, e := strconv.Atoi(v); e == nil {
					cfg.LLM.TimeoutS = n
				}
			}
		case "memory":
			switch k {
			case "backend":
				cfg.Memory.Backend = v
			case "redis_addr":
				cfg.Memory.RedisAddr = v
			case "redis_db":
				if n, e := strconv.Atoi(v); e == nil {
					cfg.Memory.RedisDB = n
				}
			case "redis_password":
				cfg.Memory.RedisPass = v
			case "key_prefix":
				cfg.Memory.KeyPrefix = v
			case "history_size":
				if n, e := strconv.Atoi(v); e == nil {
					cfg.Memory.HistorySize = n
				}
			}
		case "agent":
			switch k {
			case "system_prompt":
				cfg.Agent.SystemPrompt = v
			case "knowledge_path":
				cfg.Agent.KnowledgePath = v
			case "knowledge_top_k":
				if n, e := strconv.Atoi(v); e == nil {
					cfg.Agent.KnowledgeTopK = n
				}
			}
		}
	}
	return s.Err()
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8080"
	}
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "openai"
	}
	if cfg.LLM.BaseURL == "" {
		cfg.LLM.BaseURL = "https://api.openai.com"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "gpt-4o-mini"
	}
	if cfg.LLM.TimeoutS <= 0 {
		cfg.LLM.TimeoutS = 30
	}
	if cfg.Memory.Backend == "" {
		cfg.Memory.Backend = "in_memory"
	}
	if cfg.Memory.RedisAddr == "" {
		cfg.Memory.RedisAddr = "redis:6379"
	}
	if cfg.Memory.KeyPrefix == "" {
		cfg.Memory.KeyPrefix = "agent"
	}
	if cfg.Memory.HistorySize <= 0 {
		cfg.Memory.HistorySize = 20
	}
	if cfg.Agent.SystemPrompt == "" {
		cfg.Agent.SystemPrompt = "你是电商平台的智能客服助手。请礼貌、简洁、可执行地回答用户问题。"
	}
	if cfg.Agent.KnowledgePath == "" {
		cfg.Agent.KnowledgePath = "configs/knowledge_base.json"
	}
	if cfg.Agent.KnowledgeTopK <= 0 {
		cfg.Agent.KnowledgeTopK = 3
	}
}

func applyEnv(cfg *Config) {
	setIf := func(env string, setter func(string)) {
		if v := strings.TrimSpace(os.Getenv(env)); v != "" {
			setter(v)
		}
	}
	setIf("SERVER_ADDRESS", func(v string) { cfg.Server.Address = v })
	setIf("LLM_PROVIDER", func(v string) { cfg.LLM.Provider = v })
	setIf("LLM_BASE_URL", func(v string) { cfg.LLM.BaseURL = v })
	setIf("OPENAI_API_KEY", func(v string) { cfg.LLM.APIKey = v })
	setIf("LLM_API_KEY", func(v string) { cfg.LLM.APIKey = v })
	setIf("LLM_MODEL", func(v string) { cfg.LLM.Model = v })
	setIf("LLM_TIMEOUT_SECONDS", func(v string) {
		if n, e := strconv.Atoi(v); e == nil {
			cfg.LLM.TimeoutS = n
		}
	})
	setIf("MEMORY_BACKEND", func(v string) { cfg.Memory.Backend = v })
	setIf("REDIS_ADDR", func(v string) { cfg.Memory.RedisAddr = v })
	setIf("REDIS_DB", func(v string) {
		if n, e := strconv.Atoi(v); e == nil {
			cfg.Memory.RedisDB = n
		}
	})
	setIf("REDIS_PASSWORD", func(v string) { cfg.Memory.RedisPass = v })
	setIf("MEMORY_KEY_PREFIX", func(v string) { cfg.Memory.KeyPrefix = v })
	setIf("MEMORY_HISTORY_SIZE", func(v string) {
		if n, e := strconv.Atoi(v); e == nil {
			cfg.Memory.HistorySize = n
		}
	})
	setIf("AGENT_SYSTEM_PROMPT", func(v string) { cfg.Agent.SystemPrompt = v })
	setIf("KNOWLEDGE_PATH", func(v string) { cfg.Agent.KnowledgePath = v })
	setIf("KNOWLEDGE_TOP_K", func(v string) {
		if n, e := strconv.Atoi(v); e == nil {
			cfg.Agent.KnowledgeTopK = n
		}
	})
}
