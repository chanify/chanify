//go:build !test

package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/viper"
)

type controlledHTTPService struct {
	started, stopped chan struct{}
	once             sync.Once
	shutdownErr      error
	closed           bool
}

func (s *controlledHTTPService) ListenAndServe() error {
	close(s.started)
	<-s.stopped
	return http.ErrServerClosed
}
func (s *controlledHTTPService) Shutdown(context.Context) error {
	if s.shutdownErr == nil {
		s.once.Do(func() { close(s.stopped) })
	}
	return s.shutdownErr
}
func (s *controlledHTTPService) Close() error {
	s.closed = true
	s.once.Do(func() { close(s.stopped) })
	return nil
}

func TestHTTPServiceShutdown(t *testing.T) {
	for _, fail := range []bool{false, true} {
		s := &controlledHTTPService{started: make(chan struct{}), stopped: make(chan struct{})}
		want := errors.New("shutdown failed")
		if fail {
			s.shutdownErr = want
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() { <-s.started; cancel() }()
		err := serveHTTP(ctx, s)
		if fail {
			if !errors.Is(err, want) || !s.closed {
				t.Fatal(err, s.closed)
			}
		} else if err != nil || s.closed {
			t.Fatal(err, s.closed)
		}
	}
}

func TestServeInitializationAndListenFailure(t *testing.T) {
	resetConfig(t)
	c := commandForTest(t, "serve")
	viper.Set("server.dburl", "missing-driver://")
	if err := runServe(context.Background(), c); err == nil || !strings.Contains(err.Error(), "init service") {
		t.Fatal(err)
	}
	viper.Set("server.dburl", "sqlite://?mode=memory")
	viper.Set("server.host", "127.0.0.1")
	viper.Set("server.port", 0)
	viper.SetConfigFile("synthetic-config.yml")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c.SetContext(ctx)
	if err := c.RunE(c, nil); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	viper.Set("server.port", listener.Addr().(*net.TCPAddr).Port)
	if err := runServe(context.Background(), c); err == nil {
		t.Fatal("listen failure ignored")
	}
}
