package tools

import (
	"context"
	"fmt"
)

type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool { return &CalculatorTool{} }

func (c *CalculatorTool) Name() string { return "calculator" }

func (c *CalculatorTool) Description() string {
	return "Do simple arithmetic with fields: a(number), b(number), op(one of + - * /)."
}

func (c *CalculatorTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"a":  map[string]any{"type": "number"},
			"b":  map[string]any{"type": "number"},
			"op": map[string]any{"type": "string", "enum": []string{"+", "-", "*", "/"}},
		},
		"required": []string{"a", "b", "op"},
	}
}

func (c *CalculatorTool) Call(_ context.Context, input map[string]any) (string, error) {
	a, ok := asFloat(input["a"])
	if !ok {
		return "", fmt.Errorf("invalid a")
	}
	b, ok := asFloat(input["b"])
	if !ok {
		return "", fmt.Errorf("invalid b")
	}
	op, _ := input["op"].(string)
	switch op {
	case "+":
		return fmt.Sprintf("%g", a+b), nil
	case "-":
		return fmt.Sprintf("%g", a-b), nil
	case "*":
		return fmt.Sprintf("%g", a*b), nil
	case "/":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		return fmt.Sprintf("%g", a/b), nil
	default:
		return "", fmt.Errorf("unsupported op")
	}
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	default:
		return 0, false
	}
}
