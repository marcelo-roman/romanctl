package sampleitem

import (
	"errors"
	"testing"
	"time"

	"sample-app/internal/domain/shared"
)

func TestCreateRequiresName(t *testing.T) {
	var rule *shared.RuleError
	if _, err := Create(" ", time.Now()); !errors.As(err, &rule) {
		t.Fatalf("expected rule error, got %v", err)
	}
	e, err := Create("  Primeiro ", time.Now())
	if err != nil || e.Name() != "Primeiro" {
		t.Fatalf("%v %v", e, err)
	}
}
