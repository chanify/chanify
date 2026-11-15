package sender

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOutcomeClassification(t *testing.T) {
	for _, tc := range []struct {
		cause                 error
		before                bool
		status, code, message string
	}{
		{context.Canceled, true, "failed", "canceled", "The operation was canceled before submission."},
		{context.DeadlineExceeded, true, "failed", "timeout", "The operation timed out before submission."},
		{io.EOF, true, "failed", "internal_error", "Cannot submit the notification."},
		{&HTTPError{401}, false, "failed", "unauthorized", "The node rejected the sending credential."},
		{&HTTPError{403}, false, "failed", "unauthorized", "The node rejected the sending credential."},
		{&HTTPError{413}, false, "failed", "too_large", "The node rejected the request size."},
		{&HTTPError{500}, false, "unknown", "outcome_unknown", "The notification may have been submitted. Retrying may duplicate it."},
		{io.EOF, false, "unknown", "outcome_unknown", "The notification may have been submitted. Retrying may duplicate it."},
	} {
		var err error
		if tc.before {
			err = beforeSubmission(tc.cause)
		} else {
			err = afterSubmission(tc.cause)
		}
		var outcome *OutcomeError
		if !errors.As(err, &outcome) || outcome.Status != tc.status || outcome.Code != tc.code || err.Error() != tc.message || !errors.Is(err, tc.cause) {
			t.Fatalf("unexpected outcome: %#v", err)
		}
	}
	if got := (&HTTPError{418}).Error(); got != "node returned HTTP 418" {
		t.Fatal(got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestRequestFailures(t *testing.T) {
	for _, tc := range []struct {
		name         string
		transportErr error
		body         io.Reader
		want         string
	}{
		{"canceled", context.Canceled, nil, "context canceled"},
		{"deadline", context.DeadlineExceeded, nil, "context deadline exceeded"},
		{"network", io.EOF, nil, "node request failed; check endpoint and connectivity"},
		{"read", nil, failingReader{}, "cannot read node response"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := New("https://node.example", "test", &http.Client{Transport: testTransport(func(*http.Request) (*http.Response, error) {
				if tc.transportErr != nil {
					return nil, tc.transportErr
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(tc.body)}, nil
			})})
			if _, err := c.Send(context.Background(), nil, "", ""); err == nil || err.Error() != tc.want {
				t.Fatalf("got %v", err)
			}
		})
	}
	c, _ := New("https://node.example", "test", nil)
	c.endpoint = ":"
	if err := c.request(context.Background(), "GET", "", "", "", nil, &Info{}); err == nil || err.Error() != "cannot construct node request" {
		t.Fatal(err)
	}
}

func TestInfoAndSubmission(t *testing.T) {
	for _, tc := range []struct {
		body    string
		success bool
	}{{`{"nodeid":"node","features":["file"]}`, true}, {`{}`, false}, {`invalid`, false}} {
		c, _ := New("https://node.example/", "test", &http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.Path != "/rest/v1/info" || r.Header.Get("User-Agent") != "chanify-cli/test" {
				t.Fatalf("unexpected request: %v", r)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
		})})
		info, err := c.Info(context.Background())
		if (err == nil) != tc.success {
			t.Fatal(info, err)
		}
		if tc.success && (info.NodeID != "node" || len(info.Features) != 1) {
			t.Fatal(info)
		}
	}
	calls := 0
	c, _ := New("https://node.example", "test", &http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("token") != "synthetic" || r.Header.Get("Content-Type") != "text/plain" {
			t.Fatal(r.Header)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"res":200,"request-uid":"receipt"}`))}, nil
	})})
	result, err := c.Submit(context.Background(), strings.NewReader("message"), "text/plain", "synthetic")
	if err != nil || result.RequestID != "receipt" {
		t.Fatal(result, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = c.Submit(ctx, nil, "", ""); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("canceled request submitted: %d", calls)
	}
	c.http.Transport = testTransport(func(*http.Request) (*http.Response, error) { return nil, io.EOF })
	_, err = c.Submit(context.Background(), nil, "", "")
	var outcome *OutcomeError
	if !errors.As(err, &outcome) || outcome.Status != "unknown" {
		t.Fatal(err)
	}
}
