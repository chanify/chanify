package sender

import (
	"context"
	"errors"
	"net/http"
)

// OutcomeError preserves whether notification submission may already have happened.
type OutcomeError struct {
	Status string
	Code   string
	Cause  error
}

func (e *OutcomeError) Error() string {
	if e.Status == "unknown" {
		return "The notification may have been submitted. Retrying may duplicate it."
	}
	switch e.Code {
	case "unauthorized":
		return "The node rejected the sending credential."
	case "too_large":
		return "The node rejected the request size."
	case "timeout":
		return "The operation timed out before submission."
	case "canceled":
		return "The operation was canceled before submission."
	}
	return "Cannot submit the notification."
}
func (e *OutcomeError) Unwrap() error { return e.Cause }
func beforeSubmission(err error) error {
	code := "internal_error"
	if errors.Is(err, context.DeadlineExceeded) {
		code = "timeout"
	}
	if errors.Is(err, context.Canceled) {
		code = "canceled"
	}
	return &OutcomeError{Status: "failed", Code: code, Cause: err}
}
func afterSubmission(err error) error {
	var status *HTTPError
	if errors.As(err, &status) {
		switch status.Status {
		case http.StatusUnauthorized, http.StatusForbidden:
			return &OutcomeError{Status: "failed", Code: "unauthorized", Cause: err}
		case http.StatusRequestEntityTooLarge:
			return &OutcomeError{Status: "failed", Code: "too_large", Cause: err}
		}
	}
	return &OutcomeError{Status: "unknown", Code: "outcome_unknown", Cause: err}
}
