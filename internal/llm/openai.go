package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func NewOpenAIClient(baseURL, apiKey, model string, timeout time.Duration) *OpenAIClient {
	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []apiMsg  `json:"messages"`
	Tools    []apiTool `json:"tools,omitempty"`
	ToolMode string    `json:"tool_choice,omitempty"`
}

type apiTool struct {
	Type     string          `json:"type"`
	Function apiToolFunction `json:"function"`
}

type apiToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type apiMsg struct {
	Role       string        `json:"role"`
	Content    string        `json:"content,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []apiToolCall `json:"tool_calls,omitempty"`
}

type apiToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function apiToolFunctionCall `json:"function"`
}

type apiToolFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResp struct {
	Choices []struct {
		Message apiMsg `json:"message"`
	} `json:"choices"`
}

func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, tools []ToolDef) (Message, error) {
	apiMsgs := make([]apiMsg, 0, len(messages))
	for _, m := range messages {
		apiMsgs = append(apiMsgs, apiMsg{Role: m.Role, Content: m.Content, ToolCallID: m.ToolCallID})
	}
	apiTools := make([]apiTool, 0, len(tools))
	for _, t := range tools {
		apiTools = append(apiTools, apiTool{Type: "function", Function: apiToolFunction{Name: t.Name, Description: t.Description, Parameters: t.Schema}})
	}

	reqBody := chatReq{Model: c.model, Messages: apiMsgs}
	if len(apiTools) > 0 {
		reqBody.Tools = apiTools
		reqBody.ToolMode = "auto"
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return Message{}, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return Message{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("chat request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return Message{}, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(body))
	}
	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Message{}, fmt.Errorf("decode response: %w", err)
	}
	if len(out.Choices) == 0 {
		return Message{}, fmt.Errorf("empty choices")
	}
	m := out.Choices[0].Message
	res := Message{Role: m.Role, Content: m.Content}
	for _, tc := range m.ToolCalls {
		args := map[string]any{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		res.ToolCalls = append(res.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: args})
	}
	return res, nil
}
