# Go Agent Starter Kit v1.1（智能客服 Demo）

这是一个可运行的 **Go 智能客服 Starter Kit**，包含：

- ✅ LLM 对话编排
- ✅ Redis / In-Memory 会话记忆
- ✅ Tool 调用系统（内置 calculator）
- ✅ 知识库检索增强（RAG-lite）
- ✅ HTTP API 服务
- ✅ Docker 部署

## 目录结构

```text
go-agent-starter-kit/
├── cmd/server/main.go
├── internal/
│   ├── agent/
│   ├── config/
│   ├── knowledge/
│   ├── llm/
│   ├── memory/
│   ├── server/
│   └── tools/
├── configs/
│   ├── config.yaml
│   └── knowledge_base.json
├── docker/Dockerfile
└── README.md
```

## 启动

```bash
go run ./cmd/server
```

## 智能客服 Demo

### 1) 健康检查

```bash
curl http://localhost:8080/healthz
```

### 2) 咨询物流问题（命中知识库）

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"u1001","message":"我昨天买的东西什么时候发货？"}'
```

### 3) 咨询退款问题（命中知识库）

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"u1001","message":"订单已经发货了还能退款吗？"}'
```

### 4) 计算类问题（工具调用）

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"u1001","message":"请帮我计算 12 * 8"}'
```

## 配置说明

`configs/config.yaml` 关键项：

- `memory.backend`: `redis` / `in_memory`
- `agent.system_prompt`: 客服系统提示词
- `agent.knowledge_path`: 知识库文件路径
- `agent.knowledge_top_k`: 每次检索注入条数

可通过环境变量覆盖：

- `OPENAI_API_KEY` / `LLM_API_KEY`
- `MEMORY_BACKEND`, `REDIS_ADDR`
- `AGENT_SYSTEM_PROMPT`, `KNOWLEDGE_PATH`, `KNOWLEDGE_TOP_K`

## Docker

```bash
docker build -f docker/Dockerfile -t go-agent-starter-kit:1.1 .
docker run --rm -p 8080:8080 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -e REDIS_ADDR=host.docker.internal:6379 \
  go-agent-starter-kit:1.1
```
