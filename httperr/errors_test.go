package httperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/weprodev/go-pkg/httperr"
)

func TestServiceError_Error(t *testing.T) {
	err := httperr.NewServiceError(400, "bad input", "field x is required", "field y must be positive")
	expected := "bad input: [field x is required field y must be positive]"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}

	errNoDetails := httperr.NewServiceError(401, "unauthorized")
	if errNoDetails.Error() != "unauthorized" {
		t.Errorf("got %q, want %q", errNoDetails.Error(), "unauthorized")
	}
}

func TestServiceError_Is(t *testing.T) {
	err1 := httperr.NewServiceError(404, "not found")
	err2 := httperr.NewServiceError(404, "not found", "different details shouldn't matter")
	err3 := httperr.NewServiceError(404, "some other message")
	err4 := httperr.NewServiceError(400, "not found")

	if !errors.Is(err2, err1) {
		t.Errorf("expected err2 to be equivalent to err1")
	}
	if errors.Is(err3, err1) {
		t.Errorf("expected err3 not to be equivalent to err1 (messages differ)")
	}
	if errors.Is(err4, err1) {
		t.Errorf("expected err4 not to be equivalent to err1 (codes differ)")
	}
	if errors.Is(errors.New("standard error"), err1) {
		t.Errorf("expected standard error not to match ServiceError")
	}
}

func TestAsServiceError(t *testing.T) {
	if httperr.AsServiceError(nil) != nil {
		t.Errorf("nil error should return nil")
	}

	stdErr := errors.New("just a standard error")
	if httperr.AsServiceError(stdErr) != nil {
		t.Errorf("standard error should return nil")
	}

	se := httperr.NewServiceError(400, "bad")
	wrapped := fmt.Errorf("wrapped: %w", se)

	extracted := httperr.AsServiceError(wrapped)
	if extracted == nil {
		t.Fatalf("failed to extract ServiceError")
	}
	if extracted.Code != 400 || extracted.Message != "bad" {
		t.Errorf("extracted wrong service error")
	}
}
