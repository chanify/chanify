package sender

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestRejectRedirectsAndAmbiguousResponses(t *testing.T) {
	calls := 0
	client, err := New("https://node.example", "test", &http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 307, Header: http.Header{"Location": []string{"https://other.example/capture"}}, Body: io.NopCloser(strings.NewReader("")), Request: r}, nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Send(context.Background(), strings.NewReader("body"), "text/plain", "synthetic-token"); err == nil {
		t.Fatal("redirect accepted")
	}
	if calls != 1 {
		t.Fatal("credential-bearing redirect followed")
	}
	for _, body := range []string{`{}`, `{"request-uid":"id","res":401}`, `invalid`, strings.Repeat("x", (1<<20)+1)} {
		client, _ := New("https://node.example", "test", &http.Client{Transport: testTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
		})})
		if _, err := client.Send(context.Background(), nil, "text/plain", "synthetic-token"); err == nil {
			t.Fatal("unconfirmed submission accepted")
		}
	}
}
func TestEndpointValidation(t *testing.T) {
	for _, endpoint := range []string{"file:///tmp/file", "https://user:secret@node.example", "https://node.example/?token=secret", "https://node.example/#secret"} {
		if _, err := New(endpoint, "test", nil); err == nil {
			t.Fatal("invalid endpoint accepted")
		}
	}
	for _, endpoint := range []string{"", DefaultEndpoint, DefaultEndpoint + "/", "https://api.chanify.net:443"} {
		client, err := New(endpoint, "test", nil)
		if err != nil || !client.IsDefault() {
			t.Fatal("default endpoint not recognized", err)
		}
	}
}
