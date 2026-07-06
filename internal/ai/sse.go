package ai

import (
	"bufio"
	"io"
	"strings"
)

// sseScanner reads Server-Sent Events from an io.Reader.
type sseScanner struct {
	scanner   *bufio.Scanner
	eventType string
	data      string
}

func newSSEScanner(r io.Reader) *sseScanner {
	return &sseScanner{scanner: bufio.NewScanner(r)}
}

// Scan reads the next complete SSE event. Returns false when done.
func (s *sseScanner) Scan() bool {
	s.eventType = ""
	s.data = ""

	var dataLines []string

	for s.scanner.Scan() {
		line := s.scanner.Text()

		if line == "" {
			// Empty line = end of event
			if len(dataLines) > 0 {
				s.data = strings.Join(dataLines, "\n")
				return true
			}
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			s.eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		} else if line == "data:" {
			dataLines = append(dataLines, "")
		}
	}

	// Handle final event without trailing newline
	if len(dataLines) > 0 {
		s.data = strings.Join(dataLines, "\n")
		return true
	}

	return false
}

func (s *sseScanner) EventType() string { return s.eventType }
func (s *sseScanner) Data() string      { return s.data }
