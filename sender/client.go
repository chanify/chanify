// Package sender provides the HTTP client shared by CLI sending and stdio MCP.
package sender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultEndpoint = "https://api.chanify.net"

type Info struct {
	NodeID   string   `json:"nodeid"`
	Features []string `json:"features"`
}
type Result struct {
	RequestID string `json:"request-uid"`
	Status    int    `json:"res"`
}
type HTTPError struct{ Status int }

func (e *HTTPError) Error() string { return fmt.Sprintf("node returned HTTP %d", e.Status) }

type Client struct {
	endpoint string
	http     *http.Client
	version  string
}

func New(endpoint, version string, transport *http.Client) (*Client, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("endpoint must be an HTTP(S) URL without credentials, query or fragment")
	}
	client := http.Client{Timeout: 30 * time.Second}
	if transport != nil {
		client = *transport
		if client.Timeout == 0 {
			client.Timeout = 30 * time.Second
		}
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &Client{endpoint: strings.TrimRight(endpoint, "/"), http: &client, version: version}, nil
}
func (c *Client) request(ctx context.Context, method, path, content, token string, body io.Reader, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, body)
	if err != nil {
		return errors.New("cannot construct node request")
	}
	req.Header.Set("User-Agent", "chanify-cli/"+c.version)
	if content != "" {
		req.Header.Set("Content-Type", content)
	}
	if token != "" {
		req.Header.Set("token", token)
	}
	response, err := c.http.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return errors.New("node request failed; check endpoint and connectivity")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return &HTTPError{Status: response.StatusCode}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return errors.New("cannot read node response")
	}
	if json.Unmarshal(data, result) != nil {
		return errors.New("invalid node JSON response")
	}
	return nil
}
func (c *Client) Info(ctx context.Context) (*Info, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	info := &Info{}
	if err := c.request(ctx, "GET", "/rest/v1/info", "", "", nil, info); err != nil {
		return nil, err
	}
	if info.NodeID == "" {
		return nil, errors.New("node information is missing nodeid")
	}
	return info, nil
}
func (c *Client) Send(ctx context.Context, body io.Reader, content, token string) (*Result, error) {
	result := &Result{}
	if err := c.request(ctx, "POST", "/v1/sender", content, token, body, result); err != nil {
		return nil, err
	}
	if result.Status != 0 && result.Status != 200 {
		return nil, fmt.Errorf("node rejected notification (code %d)", result.Status)
	}
	if result.RequestID == "" {
		return nil, errors.New("node did not confirm submission; retrying may duplicate the notification")
	}
	return result, nil
}

// Submit provides classified outcomes for automated callers without changing Send's errors.
func (c *Client) Submit(ctx context.Context, body io.Reader, content, token string) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, beforeSubmission(err)
	}
	result, err := c.Send(ctx, body, content, token)
	if err != nil {
		return nil, afterSubmission(err)
	}
	return result, nil
}

func (c *Client) IsDefault() bool {
	u, _ := url.Parse(c.endpoint)
	return u.Scheme == "https" && strings.EqualFold(u.Hostname(), "api.chanify.net") && (u.Port() == "" || u.Port() == "443") && (u.Path == "" || u.Path == "/")
}
