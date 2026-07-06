package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const maxInvestigationSteps = 8

type Investigator struct {
	provider LLMProvider
	tools    *ToolRegistry
	logger   zerolog.Logger
}

func NewInvestigator(provider LLMProvider, tools *ToolRegistry, logger zerolog.Logger) *Investigator {
	return &Investigator{provider: provider, tools: tools, logger: logger}
}

type InvestigationRequest struct {
	Question string `json:"question"`
	Stream   bool   `json:"stream"`
}

type InvestigationStep struct {
	StepNumber int       `json:"step_number"`
	Type       string    `json:"type"` // "thinking", "tool_call", "tool_result", "answer"
	Content    string    `json:"content"`
	ToolName   string    `json:"tool_name,omitempty"`
	ToolArgs   string    `json:"tool_args,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

type InvestigationResponse struct {
	Answer     string              `json:"answer"`
	Confidence float64             `json:"confidence"`
	Steps      []InvestigationStep `json:"steps"`
	Usage      Usage               `json:"usage"`
}

const investigatorSystem = `You are AtlasDB Investigator, an autonomous agent that investigates incidents in a distributed event streaming platform.

Current time: %s

Your task: Investigate the user's question by gathering data using the available tools, then provide a thorough analysis.

Strategy:
1. Break down the problem into investigable components
2. Gather relevant data using tools (check metrics, search events, look at alerts)
3. Correlate findings across multiple data sources
4. Form a hypothesis about root cause
5. Provide a clear, evidence-based answer

Always cite specific data points (counts, timestamps, event IDs) in your answer.
After your investigation, rate your confidence from 0.0 to 1.0.

Be systematic: check error rates, recent events, alert history, and time-series data before concluding.`

func (inv *Investigator) Investigate(ctx context.Context, req InvestigationRequest) (*InvestigationResponse, error) {
	messages := []Message{
		{Role: RoleSystem, Content: fmt.Sprintf(investigatorSystem, time.Now().UTC().Format(time.RFC3339))},
		{Role: RoleUser, Content: req.Question},
	}

	toolDefs := inv.tools.Definitions()

	var steps []InvestigationStep
	totalUsage := Usage{}
	stepNum := 0

	for stepNum < maxInvestigationSteps {
		stepNum++

		resp, err := inv.provider.Complete(ctx, CompletionRequest{
			Messages:    messages,
			Tools:       toolDefs,
			MaxTokens:   4096,
			Temperature: 0.2,
		})
		if err != nil {
			return nil, fmt.Errorf("investigation step %d: %w", stepNum, err)
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		// If there are tool calls, execute them
		if len(resp.ToolCalls) > 0 {
			// Add assistant message with tool calls to history
			assistantMsg := Message{
				Role:      RoleAssistant,
				Content:   resp.Content,
				ToolCalls: resp.ToolCalls,
			}
			messages = append(messages, assistantMsg)

			if resp.Content != "" {
				steps = append(steps, InvestigationStep{
					StepNumber: stepNum,
					Type:       "thinking",
					Content:    resp.Content,
				})
			}

			for _, tc := range resp.ToolCalls {
				steps = append(steps, InvestigationStep{
					StepNumber: stepNum,
					Type:       "tool_call",
					ToolName:   tc.Name,
					ToolArgs:   tc.Arguments,
				})

				inv.logger.Debug().
					Str("tool", tc.Name).
					Int("step", stepNum).
					Msg("Investigation tool call")

				start := time.Now()
				result, err := inv.tools.Execute(ctx, tc.Name, tc.Arguments)
				duration := time.Since(start)

				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}

				// Truncate very large results
				if len(result) > 6000 {
					result = result[:6000] + "\n... (truncated)"
				}

				steps = append(steps, InvestigationStep{
					StepNumber: stepNum,
					Type:       "tool_result",
					ToolName:   tc.Name,
					Content:    result,
					Duration:   duration,
				})

				messages = append(messages, Message{
					Role:       RoleTool,
					Content:    result,
					ToolCallID: tc.ID,
				})
			}

			continue
		}

		// No tool calls — this is the final answer
		steps = append(steps, InvestigationStep{
			StepNumber: stepNum,
			Type:       "answer",
			Content:    resp.Content,
		})

		// Extract confidence
		confidence := extractConfidence(resp.Content)

		return &InvestigationResponse{
			Answer:     resp.Content,
			Confidence: confidence,
			Steps:      steps,
			Usage:      totalUsage,
		}, nil
	}

	// Ran out of steps — ask for final summary
	messages = append(messages, Message{
		Role:    RoleUser,
		Content: "You've reached the maximum investigation steps. Please provide your best answer now based on what you've found.",
	})

	resp, err := inv.provider.Complete(ctx, CompletionRequest{
		Messages:    messages,
		MaxTokens:   4096,
		Temperature: 0.2,
	})
	if err != nil {
		return nil, err
	}

	totalUsage.PromptTokens += resp.Usage.PromptTokens
	totalUsage.CompletionTokens += resp.Usage.CompletionTokens
	totalUsage.TotalTokens += resp.Usage.TotalTokens

	steps = append(steps, InvestigationStep{
		StepNumber: stepNum + 1,
		Type:       "answer",
		Content:    resp.Content,
	})

	return &InvestigationResponse{
		Answer:     resp.Content,
		Confidence: extractConfidence(resp.Content),
		Steps:      steps,
		Usage:      totalUsage,
	}, nil
}

// InvestigateStream runs the investigation and streams steps via a channel.
func (inv *Investigator) InvestigateStream(ctx context.Context, req InvestigationRequest) (<-chan InvestigationStep, error) {
	ch := make(chan InvestigationStep, 32)

	go func() {
		defer close(ch)

		result, err := inv.Investigate(ctx, req)
		if err != nil {
			ch <- InvestigationStep{
				Type:    "answer",
				Content: fmt.Sprintf("Investigation failed: %v", err),
			}
			return
		}

		for _, step := range result.Steps {
			select {
			case <-ctx.Done():
				return
			case ch <- step:
			}
		}
	}()

	return ch, nil
}

func extractConfidence(text string) float64 {
	// Try to parse confidence from structured output
	lower := strings.ToLower(text)

	// Look for "confidence: X.X" or similar patterns
	var confidence float64
	for _, prefix := range []string{`"confidence":`, `"confidence" :`, `confidence:`} {
		idx := strings.Index(lower, prefix)
		if idx >= 0 {
			remaining := strings.TrimSpace(lower[idx+len(prefix):])
			if _, err := fmt.Sscanf(remaining, "%f", &confidence); err == nil {
				if confidence >= 0 && confidence <= 1.0 {
					return confidence
				}
			}
		}
	}

	// Try JSON parse
	type confJSON struct {
		Confidence float64 `json:"confidence"`
	}
	var cj confJSON
	if err := json.Unmarshal([]byte(text), &cj); err == nil && cj.Confidence > 0 {
		return cj.Confidence
	}

	return 0.7 // default moderate confidence
}
