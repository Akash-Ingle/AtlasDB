package ingest

import (
	"fmt"

	"github.com/atlasdb/atlasdb/pkg/models"
)

type ValidationError struct {
	Index   int    `json:"index"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("event[%d].%s: %s", e.Index, e.Field, e.Message)
}

func ValidateEventInput(idx int, e models.EventInput, maxSizeBytes int) []ValidationError {
	var errs []ValidationError

	if e.Source == "" {
		errs = append(errs, ValidationError{Index: idx, Field: "source", Message: "required"})
	} else if len(e.Source) > 256 {
		errs = append(errs, ValidationError{Index: idx, Field: "source", Message: "must be <= 256 characters"})
	}

	if e.EventType == "" {
		errs = append(errs, ValidationError{Index: idx, Field: "event_type", Message: "required"})
	} else if len(e.EventType) > 256 {
		errs = append(errs, ValidationError{Index: idx, Field: "event_type", Message: "must be <= 256 characters"})
	}

	if e.Severity != "" && !e.Severity.Valid() {
		errs = append(errs, ValidationError{Index: idx, Field: "severity", Message: "must be one of: debug, info, warn, error, fatal"})
	}

	if len(e.Data) > maxSizeBytes {
		errs = append(errs, ValidationError{Index: idx, Field: "data", Message: fmt.Sprintf("must be <= %d bytes", maxSizeBytes)})
	}

	if len(e.Tags) > 50 {
		errs = append(errs, ValidationError{Index: idx, Field: "tags", Message: "must have <= 50 tags"})
	}

	return errs
}
