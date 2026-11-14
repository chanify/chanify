package mcpnotify

import (
	"context"
	"errors"

	"github.com/chanify/chanify/sender"
)

type Submission struct {
	Format    string `json:"format"`
	RequestID string `json:"request_id"`
}
type SendError struct {
	Status  string
	Code    string
	Message string
	Cause   error
}

func (e *SendError) Error() string { return e.Message }
func (e *SendError) Unwrap() error { return e.Cause }
func rejected(code, message string) error {
	return &SendError{Status: "failed", Code: code, Message: message}
}
func preflightError(err error) error {
	code, message := "discovery_failed", "Cannot discover node capabilities."
	var status *sender.HTTPError
	if errors.Is(err, context.DeadlineExceeded) {
		code, message = "timeout", "The operation timed out before submission."
	} else if errors.Is(err, context.Canceled) {
		code, message = "canceled", "The operation was canceled before submission."
	} else if errors.As(err, &status) && (status.Status == 401 || status.Status == 403) {
		code, message = "unauthorized", "The node rejected access."
	}
	return &SendError{Status: "failed", Code: code, Message: message, Cause: err}
}
func sendingError(err error) error {
	var outcome *sender.OutcomeError
	if errors.As(err, &outcome) {
		return &SendError{Status: outcome.Status, Code: outcome.Code, Message: outcome.Error(), Cause: err}
	}
	return &SendError{Status: "unknown", Code: "outcome_unknown", Message: "The notification may have been submitted. Retrying may duplicate it.", Cause: err}
}
