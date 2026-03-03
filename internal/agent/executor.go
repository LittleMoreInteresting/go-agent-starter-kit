package agent

import (
	"context"
	"fmt"

	"go-agent-starter-kit/internal/llm"
	"go-agent-starter-kit/internal/tools"
)

type Executor struct {
	registry *tools.Registry
}

func NewExecutor(registry *tools.Registry) *Executor {
	return &Executor{registry: registry}
}

func (e *Executor) ToolDefs() []llm.ToolDef {
	defs := make([]llm.ToolDef, 0)
	for _, t := range e.registry.List() {
		defs = append(defs, llm.ToolDef{Name: t.Name(), Description: t.Description(), Schema: t.InputSchema()})
	}
	return defs
}

func (e *Executor) RunCalls(ctx context.Context, calls []llm.ToolCall) []llm.Message {
	results := make([]llm.Message, 0, len(calls))
	for _, c := range calls {
		tool, err := e.registry.Get(c.Name)
		if err != nil {
			results = append(results, llm.Message{Role: "tool", ToolCallID: c.ID, Content: err.Error()})
			continue
		}
		out, err := tool.Call(ctx, c.Arguments)
		if err != nil {
			out = fmt.Sprintf("tool error: %v", err)
		}
		results = append(results, llm.Message{Role: "tool", ToolCallID: c.ID, Content: out})
	}
	return results
}
