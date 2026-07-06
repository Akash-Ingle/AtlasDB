package search

import (
	"fmt"
	"strings"
)

// ParsedQuery represents a parsed search query with field-scoped terms and free-text.
//
// Syntax:
//
//	source:payment-service error       → field filter + keyword
//	severity:error AND timeout         → field filter + boolean + keyword
//	"connection refused"               → phrase search
//	source:auth-service OR source:gateway → multiple field filters
type ParsedQuery struct {
	FreeText string
	Filters  map[string]string
}

var searchableFields = map[string]string{
	"source":     "source",
	"type":       "event_type",
	"event_type": "event_type",
	"severity":   "severity",
	"status":     "data->>'status'",
	"method":     "data->>'method'",
	"path":       "data->>'path'",
	"host":       "metadata->>'host'",
}

func Parse(raw string) ParsedQuery {
	pq := ParsedQuery{
		Filters: make(map[string]string),
	}

	tokens := tokenize(raw)
	var freeTokens []string

	for _, tok := range tokens {
		upper := strings.ToUpper(tok)
		if upper == "AND" || upper == "OR" || upper == "NOT" {
			freeTokens = append(freeTokens, mapBoolean(upper))
			continue
		}

		if idx := strings.Index(tok, ":"); idx > 0 && idx < len(tok)-1 {
			field := strings.ToLower(tok[:idx])
			value := tok[idx+1:]
			value = strings.Trim(value, "\"'")

			if _, ok := searchableFields[field]; ok {
				pq.Filters[field] = value
				continue
			}
		}

		freeTokens = append(freeTokens, tok)
	}

	pq.FreeText = strings.TrimSpace(strings.Join(freeTokens, " "))
	return pq
}

// BuildSQL turns a ParsedQuery into WHERE conditions and arguments.
func (pq ParsedQuery) BuildSQL(argOffset int) (conditions []string, args []interface{}, nextArg int) {
	idx := argOffset

	for field, value := range pq.Filters {
		col, ok := searchableFields[field]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = $%d", col, idx))
		args = append(args, value)
		idx++
	}

	if pq.FreeText != "" {
		tsquery := toTSQuery(pq.FreeText)
		conditions = append(conditions, fmt.Sprintf("search_vector @@ to_tsquery('english', $%d)", idx))
		args = append(args, tsquery)
		idx++
	}

	return conditions, args, idx
}

// toTSQuery converts free text with optional AND/OR/NOT into a tsquery string.
func toTSQuery(text string) string {
	tokens := strings.Fields(text)
	var parts []string

	for i, tok := range tokens {
		switch strings.ToUpper(tok) {
		case "&":
			parts = append(parts, "&")
		case "|":
			parts = append(parts, "|")
		case "!":
			parts = append(parts, "!")
		default:
			if i > 0 && len(parts) > 0 {
				last := parts[len(parts)-1]
				if last != "&" && last != "|" && last != "!" {
					parts = append(parts, "&")
				}
			}
			tok = strings.Trim(tok, "\"'")
			parts = append(parts, sanitizeTSToken(tok))
		}
	}

	return strings.Join(parts, " ")
}

func mapBoolean(op string) string {
	switch op {
	case "AND":
		return "&"
	case "OR":
		return "|"
	case "NOT":
		return "!"
	default:
		return op
	}
}

func tokenize(input string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(input); i++ {
		c := input[i]

		if inQuotes {
			if c == quoteChar {
				current.WriteByte(c)
				tokens = append(tokens, current.String())
				current.Reset()
				inQuotes = false
			} else {
				current.WriteByte(c)
			}
			continue
		}

		if c == '"' || c == '\'' {
			inQuotes = true
			quoteChar = c
			current.WriteByte(c)
			continue
		}

		if c == ' ' || c == '\t' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func sanitizeTSToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	result := b.String()
	if result == "" {
		return "unknown"
	}
	return result
}
