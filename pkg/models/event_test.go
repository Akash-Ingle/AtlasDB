package models

import "testing"

func TestSeverityValid(t *testing.T) {
	valid := []Severity{SeverityDebug, SeverityInfo, SeverityWarn, SeverityError, SeverityFatal}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("Severity(%q).Valid() = false, want true", s)
		}
	}

	invalid := []Severity{"critical", "unknown", ""}
	for _, s := range invalid {
		if s.Valid() {
			t.Errorf("Severity(%q).Valid() = true, want false", s)
		}
	}
}
