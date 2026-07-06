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

type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	embedModel string
	httpClient *http.Client
	logger     zerolog.Logger
}

type OpenAIConfig struct {
	APIKey     string
	BaseURL    string // defaults to https://api.openai.com/v1
	Model      string // defaults to gpt-4o
	EmbedModel string // defaults to text-embedding-3-small
}

func NewOpenAIProvider(cfg OpenAIConfig, logger zerolog.Logger) *OpenAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}
	if cfg.EmbedModel == "" {
		cfg.EmbedModel = "text-embedding-3-small"
	}

	return &OpenAIProvider{
		apiKey:     cfg.APIKey,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.Model,
		embedModel: cfg.EmbedModel,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		logger:     logger,
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

// --- Complete (non-streaming) ---

type oaiRequest struct {
	Model       string       `json:"model"`
	Messages    []oaiMessage `json:"messages"`
	Tools       []oaiTool    `json:"tools,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature *float64     `json:"temperature,omitempty"`
	ResponseFormat *oaiResponseFormat `json:"response_format,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

type oaiMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  []oaiToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

type oaiTool struct {
	Type     string      `json:"type"`
	Function oaiFunction `json:"function"`
}

type oaiFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type oaiToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaiResponseFormat struct {
	Type string `json:"type"`
}

type oaiResponse struct {
	Choices []struct {
		Message      oaiMessage `json:"message"`
		Delta        oaiMessage `json:"delta"`
		FinishReason string     `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	oaiReq := p.buildRequest(req)
	oaiReq.Stream = false

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("openai returned %d: %s", resp.StatusCode, string(respBody))
	}

	var oaiResp oaiResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, fmt.Errorf("openai error: %s", oaiResp.Error.Message)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	choice := oaiResp.Choices[0]
	result := &CompletionResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: Usage{
			PromptTokens:     oaiResp.Usage.PromptTokens,
			CompletionTokens: oaiResp.Usage.CompletionTokens,
			TotalTokens:      oaiResp.Usage.TotalTokens,
		},
		Latency: time.Since(start),
	}

	for _, tc := range choice.Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return result, nil
}

// --- CompleteStream ---

func (p *OpenAIProvider) CompleteStream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	oaiReq := p.buildRequest(req)
	oaiReq.Stream = true

	body, err := json.Marshal(oaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai stream request: %w", err)
	}

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("openai returned %d: %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamChunk, 64)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := newSSEScanner(resp.Body)
		for scanner.Scan() {
			data := scanner.Data()
			if data == "[DONE]" {
				ch <- StreamChunk{Done: true}
				return
			}

			var oaiResp oaiResponse
			if err := json.Unmarshal([]byte(data), &oaiResp); err != nil {
				continue
			}

			if len(oaiResp.Choices) == 0 {
				continue
			}

			delta := oaiResp.Choices[0].Delta
			chunk := StreamChunk{
				Content:      delta.Content,
				FinishReason: oaiResp.Choices[0].FinishReason,
			}

			for _, tc := range delta.ToolCalls {
				chunk.ToolCalls = append(chunk.ToolCalls, ToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				})
			}

			ch <- chunk
		}
	}()

	return ch, nil
}

// --- Embed ---

type oaiEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type oaiEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody, _ := json.Marshal(oaiEmbedRequest{
		Model: p.embedModel,
		Input: texts,
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embed request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("embed returned %d: %s", resp.StatusCode, string(body))
	}

	var embedResp oaiEmbedResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("parse embed response: %w", err)
	}

	if embedResp.Error != nil {
		return nil, fmt.Errorf("embed error: %s", embedResp.Error.Message)
	}

	result := make([][]float32, len(embedResp.Data))
	for i, d := range embedResp.Data {
		result[i] = d.Embedding
	}
	return result, nil
}

// --- Helpers ---

func (p *OpenAIProvider) buildRequest(req CompletionRequest) oaiRequest {
	model := req.Model
	if model == "" {
		model = p.model
	}

	oai := oaiRequest{
		Model:     model,
		MaxTokens: req.MaxTokens,
	}

	if req.Temperature != 0 {
		t := req.Temperature
		oai.Temperature = &t
	}

	if req.JSONMode {
		oai.ResponseFormat = &oaiResponseFormat{Type: "json_object"}
	}

	for _, m := range req.Messages {
		om := oaiMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			om.ToolCalls = append(om.ToolCalls, oaiToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tc.Name, Arguments: tc.Arguments},
			})
		}
		oai.Messages = append(oai.Messages, om)
	}

	for _, t := range req.Tools {
		oai.Tools = append(oai.Tools, oaiTool{
			Type: "function",
			Function: oaiFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	return oai
}
