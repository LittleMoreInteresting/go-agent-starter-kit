# Go Agent Starter Kit v1.0

生产级 Go Agent Starter Kit，包含：

- ✅ 可运行完整代码
- ✅ Redis Memory（可切换 In-Memory）
- ✅ Tool 调用系统（内置 calculator 示例）
- ✅ HTTP API（`/v1/chat`）
- ✅ Docker 部署

## 目录结构

```text
go-agent-starter-kit/
├── cmd/server/main.go
├── internal/
│   ├── agent/
│   ├── config/
│   ├── llm/
│   ├── memory/
│   ├── server/
│   └── tools/
├── pkg/logger/logger.go
├── configs/config.yaml
├── docker/Dockerfile
├── go.mod
└── README.md
```

## 快速启动

### 1) 本地运行

```bash
go mod tidy
go run ./cmd/server
```

### 2) 配置

配置文件在 `configs/config.yaml`，也可通过环境变量覆盖：

- `OPENAI_API_KEY` / `LLM_API_KEY`
- `LLM_MODEL`
- `LLM_BASE_URL`
- `MEMORY_BACKEND` (`redis` / `in_memory`)
- `REDIS_ADDR`
- `SERVER_ADDRESS`

### 3) 调用 API

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"demo","message":"计算 12 * 8"}'
```

### 4) 健康检查

```bash
curl http://localhost:8080/healthz
```

## Docker 部署

```bash
docker build -f docker/Dockerfile -t go-agent-starter-kit:1.0 .
docker run --rm -p 8080:8080 \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -e REDIS_ADDR=host.docker.internal:6379 \
  go-agent-starter-kit:1.0
```

## Agent 工作流

1. 用户消息写入 Memory（Redis 或 In-Memory）
2. 读取会话历史并发送给 LLM
3. 若 LLM 返回 Tool Calls，执行工具
4. 将 Tool 输出回填给 LLM 二次生成最终回复
5. 回复写入 Memory

## 内置 Tool

- `calculator`：
  - 入参：`a`、`b`、`op(+,-,*,/)`
  - 返回运算结果字符串

## 生产建议

- 在网关层增加鉴权（JWT / API Key）
- 增加请求限流和审计日志
- Redis 开启持久化与高可用部署
- 给 LLM 和 Tool 调用增加 tracing（OpenTelemetry）
- 按业务场景扩展更多工具与安全策略
