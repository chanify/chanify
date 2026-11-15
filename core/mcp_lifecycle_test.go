package core

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chanify/chanify/crypto"
	"github.com/chanify/chanify/mcpnotify"
)

func TestMCPFactoryAndStoppedRequests(t *testing.T) {
	c, raw, _ := mcpFixture(t, false, false)
	factory := c.mcpServerFactory(false)
	r := httptest.NewRequest("POST", "/mcp", nil)
	if factory(r) != nil {
		t.Fatal("server created without authorization")
	}
	record, err := c.logic.AuthenticateMCP(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(r.Context(), mcpAuthKey{}, &mcpAuthorization{raw: raw, record: record})
	a, b := factory(r.WithContext(ctx)), factory(r.WithContext(ctx))
	if a == nil || a != b {
		t.Fatal("credential server was not reused")
	}
	h := c.initMCP(false)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || w.Header().Get("WWW-Authenticate") == "" {
		t.Fatal(w.Code)
	}
	c.Close()
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}

func TestMCPMonitorHeartbeatAndRevocation(t *testing.T) {
	c, _ := persistentCore(t)
	key := crypto.GenerateSecretKey([]byte("synthetic-mcp-account"))
	record, raw, err := c.logic.CreateMCPToken(key.ToID(0), "Agent", time.Now().Add(time.Hour).Unix())
	if err != nil {
		t.Fatal(err)
	}
	id := record.ID
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticks := make(chan time.Time)
	heartbeat := make(chan struct{}, 1)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		c.monitorMCP(ctx, nil, ticks, raw, true, func() { heartbeat <- struct{}{} }, cancel)
	}()
	for i := 0; i < 15; i++ {
		select {
		case ticks <- time.Now():
		case <-ctx.Done():
			t.Fatal("monitor did not accept ticks")
		}
	}
	select {
	case <-heartbeat:
	case <-ctx.Done():
		t.Fatal("heartbeat missing")
	}
	if err := c.logic.UpdateMCPToken(record.UserID, id, "", true); err != nil {
		t.Fatal(err)
	}
	select {
	case ticks <- time.Now():
	case <-ctx.Done():
	}
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("monitor did not stop after revocation")
	}
	if ctx.Err() != context.Canceled {
		t.Fatal("revocation did not cancel the request", ctx.Err())
	}
}

func TestMCPMonitorStopPaths(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.monitorMCP(ctx, nil, nil, "", false, func() { t.Fatal("unexpected heartbeat") }, func() { t.Fatal("unexpected cancellation") })
	done := make(chan struct{})
	close(done)
	c.monitorMCP(context.Background(), done, nil, "", false, func() { t.Fatal("unexpected heartbeat") }, func() { t.Fatal("unexpected cancellation") })
}

func TestMCPPreparationAndSubmissionGuards(t *testing.T) {
	c, _, _ := mcpFixture(t, false, false)
	for _, tc := range []struct {
		name string
		args *mcpnotify.Arguments
		code string
	}{
		{"send_text", &mcpnotify.Arguments{Channel: "sys.device", Text: "text"}, "invalid_channel"},
		{"send_text", &mcpnotify.Arguments{Channel: "sys.none", Text: strings.Repeat("x", 2000)}, "unsupported"},
		{"send_file", &mcpnotify.Arguments{Channel: "sys.none", Filename: "report.txt", Bytes: []byte("data")}, "unsupported"},
	} {
		msg, _, failure := c.prepareMCPMessage(tc.name, tc.args)
		if msg != nil || failure == nil || failure.StructuredContent.(map[string]interface{})["code"] != tc.code {
			t.Fatal(msg, failure)
		}
	}
	result, err := c.submitMCPMessage(context.Background(), nil, nil, "")
	if err != nil || !result.IsError || result.StructuredContent.(map[string]interface{})["code"] != "unauthorized" {
		t.Fatal(result, err)
	}
}

func TestMCPShutdownKeepsStorageUntilClose(t *testing.T) {
	c, raw, _ := mcpFixture(t, false, true)
	c.StopMCP()
	if _, err := c.logic.AuthenticateMCP(raw); err != nil {
		t.Fatal("HTTP shutdown closed persistence too early", err)
	}
	w := httptest.NewRecorder()
	c.initMCP(true).ServeHTTP(w, httptest.NewRequest("POST", "/mcp", nil))
	if w.Code != 503 {
		t.Fatal("new MCP request accepted during shutdown", w.Code)
	}
	c.StopMCP()
}
