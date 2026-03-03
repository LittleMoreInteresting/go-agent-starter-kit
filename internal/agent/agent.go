package agent

import (
	"context"
	"fmt"
	"strings"

	"go-agent-starter-kit/internal/knowledge"
	"go-agent-starter-kit/internal/llm"
	"go-agent-starter-kit/internal/memory"
)

type Agent struct {
	llm          llm.Client
	memory       memory.Store
	exec         *Executor
	historySize  int
	systemPrompt string
	kb           knowledge.Base
	knowledgeTop int
}

func New(llmClient llm.Client, mem memory.Store, exec *Executor, historySize int, systemPrompt string, kb knowledge.Base, knowledgeTop int) *Agent {
	return &Agent{
		llm:          llmClient,
		memory:       mem,
		exec:         exec,
		historySize:  historySize,
		systemPrompt: systemPrompt,
		kb:           kb,
		knowledgeTop: knowledgeTop,
	}
}

func (a *Agent) Chat(ctx context.Context, sessionID, userInput string) (string, error) {
	if err := a.memory.Append(ctx, sessionID, memory.Message{Role: "user", Content: userInput}); err != nil {
		return "", fmt.Errorf("append user message: %w", err)
	}
	history, err := a.memory.History(ctx, sessionID, a.historySize)
	if err != nil {
		return "", fmt.Errorf("load history: %w", err)
	}
	msgs := a.buildMessages(userInput, history)

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

func (a *Agent) buildMessages(query string, history []memory.Message) []llm.Message {
	system := strings.TrimSpace(a.systemPrompt)
	if a.kb != nil && a.knowledgeTop > 0 {
		items := a.kb.Search(query, a.knowledgeTop)
		if len(items) > 0 {
			system += "\n\n可用知识库："
			for _, item := range items {
				system += "\n- [" + item.Doc.Title + "] " + item.Doc.Content
			}
			system += "\n请优先基于以上知识库回答；若知识库没有明确答案，要明确说明并给出下一步建议。"
		}
	}
	msgs := []llm.Message{{Role: "system", Content: system}}
	msgs = append(msgs, toLLMMessages(history)...)
	return msgs
}

func toLLMMessages(in []memory.Message) []llm.Message {
	out := make([]llm.Message, 0, len(in))
	for _, m := range in {
		out = append(out, llm.Message{Role: m.Role, Content: m.Content})
	}
	return out
}
