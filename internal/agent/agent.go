package agent

import (
	"context"
	"fmt"

	"go-agent-starter-kit/internal/llm"
	"go-agent-starter-kit/internal/memory"
)

type Agent struct {
	llm         llm.Client
	memory      memory.Store
	exec        *Executor
	historySize int
}

func New(llmClient llm.Client, mem memory.Store, exec *Executor, historySize int) *Agent {
	return &Agent{llm: llmClient, memory: mem, exec: exec, historySize: historySize}
}

func (a *Agent) Chat(ctx context.Context, sessionID, userInput string) (string, error) {
	if err := a.memory.Append(ctx, sessionID, memory.Message{Role: "user", Content: userInput}); err != nil {
		return "", fmt.Errorf("append user message: %w", err)
	}
	history, err := a.memory.History(ctx, sessionID, a.historySize)
	if err != nil {
		return "", fmt.Errorf("load history: %w", err)
	}
	msgs := toLLMMessages(history)

	resp, err := a.llm.Chat(ctx, msgs, a.exec.ToolDefs())
	if err != nil {
		return "", fmt.Errorf("llm chat: %w", err)
	}

	if len(resp.ToolCalls) > 0 {
		msgs = append(msgs, resp)
		toolResults := a.exec.RunCalls(ctx, resp.ToolCalls)
		msgs = append(msgs, toolResults...)
		resp, err = a.llm.Chat(ctx, msgs, a.exec.ToolDefs())
		if err != nil {
			return "", fmt.Errorf("llm chat after tools: %w", err)
		}
	}

	if err := a.memory.Append(ctx, sessionID, memory.Message{Role: "assistant", Content: resp.Content}); err != nil {
		return "", fmt.Errorf("append assistant message: %w", err)
	}
	return resp.Content, nil
}

func toLLMMessages(in []memory.Message) []llm.Message {
	out := make([]llm.Message, 0, len(in))
	for _, m := range in {
		out = append(out, llm.Message{Role: m.Role, Content: m.Content})
	}
	return out
}
