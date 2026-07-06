package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type AnthropicProvider struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
	logger     zerolog.Logger
}

type AnthropicConfig struct {
	APIKey  string
	BaseURL string
	Model   string // defaults to claude-sonnet-4-20250514
}

func NewAnthropicProvider(cfg AnthropicConfig, logger zerolog.Logger) *AnthropicProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-sonnet-4-20250514"
	}

	return &AnthropicProvider{
		apiKey:     cfg.APIKey,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.Model,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		logger:     logger,
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

// --- Anthropic API types ---

type claudeRequest struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    string         `json:"system,omitempty"`
	Messages  []claudeMsg    `json:"messages"`
	Tools     []claudeTool   `json:"tools,omitempty"`
	Stream    bool           `json:"stream,omitempty"`
}

type claudeMsg struct {
	Role    string        `json:"role"`
	Content interface{}   `json:"content"`
}

type claudeContent struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

type claudeTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type claudeResponse struct {
	Content      []claudeContent `json:"content"`
	StopReason   string          `json:"stop_reason"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	claudeReq := p.buildRequest(req)
	claudeReq.Stream = false

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, string(respBody))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s", claudeResp.Error.Message)
	}

	result := &CompletionResponse{
		FinishReason: claudeResp.StopReason,
		Usage: Usage{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
		Latency: time.Since(start),
	}

	for _, block := range claudeResp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		}
	}

	return result, nil
}

// --- CompleteStream ---

type claudeStreamEvent struct {
	Type  string          `json:"type"`
	Delta json.RawMessage `json:"delta,omitempty"`
	ContentBlock *claudeContent `json:"content_block,omitempty"`
	Message *claudeResponse `json:"message,omitempty"`
	Index int `json:"index,omitempty"`
}

func (p *AnthropicProvider) CompleteStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	claudeReq := p.buildRequest(req)
	claudeReq.Stream = true

	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic stream: %w", err)
	}

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic returned %d: %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := newSSEScanner(resp.Body)
		var currentToolCall *ToolCall

		for scanner.Scan() {
			eventType := scanner.EventType()
			data := scanner.Data()
			if data == "" {
				continue
			}

			switch eventType {
			case "content_block_start":
				var evt claudeStreamEvent
				if err := json.Unmarshal([]byte(data), &evt); err != nil {
					continue
				}
				if evt.ContentBlock != nil && evt.ContentBlock.Type == "tool_use" {
					currentToolCall = &ToolCall{
						ID:   evt.ContentBlock.ID,
						Name: evt.ContentBlock.Name,
					}
				}

			case "content_block_delta":
				var deltaWrapper struct {
					Delta struct {
						Type        string `json:"type"`
						Text        string `json:"text"`
						PartialJSON string `json:"partial_json"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &deltaWrapper); err != nil {
					continue
				}

				switch deltaWrapper.Delta.Type {
				case "text_delta":
					ch <- StreamChunk{Content: deltaWrapper.Delta.Text}
				case "input_json_delta":
					if currentToolCall != nil {
						currentToolCall.Arguments += deltaWrapper.Delta.PartialJSON
					}
				}

			case "content_block_stop":
				if currentToolCall != nil {
					ch <- StreamChunk{
						ToolCalls: []ToolCall{*currentToolCall},
					}
					currentToolCall = nil
				}

			case "message_stop":
				ch <- StreamChunk{Done: true}
				return

			case "message_delta":
				var deltaWrapper struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &deltaWrapper); err == nil {
					if deltaWrapper.Delta.StopReason != "" {
						ch <- StreamChunk{FinishReason: deltaWrapper.Delta.StopReason}
					}
				}
			}
		}
	}()

	return ch, nil
}

// Embed uses a compatible embedding API or returns an error
func (p *AnthropicProvider) Embed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, fmt.Errorf("anthropic does not natively support embeddings; use OpenAI or a local model")
}

// --- Helpers ---

func (p *AnthropicProvider) buildRequest(req CompletionRequest) claudeRequest {
	model := req.Model
	if model == "" {
		model = p.model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	cr := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
	}

	// Extract system message and convert the rest
	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			cr.System = m.Content
			continue
		}

		msg := claudeMsg{Role: string(m.Role)}

		if m.Role == RoleTool {
			// Anthropic tool results use content blocks
			msg.Role = "user"
			msg.Content = []claudeContent{{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
			}}
		} else if len(m.ToolCalls) > 0 {
			// Assistant message with tool calls
			var blocks []claudeContent
			if m.Content != "" {
				blocks = append(blocks, claudeContent{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, claudeContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: json.RawMessage(tc.Arguments),
				})
			}
			msg.Content = blocks
		} else {
			msg.Content = m.Content
		}

		cr.Messages = append(cr.Messages, msg)
	}

	for _, t := range req.Tools {
		cr.Tools = append(cr.Tools, claudeTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	return cr
}
