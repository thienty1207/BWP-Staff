package main

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/thienty1207/BWP-Staff/backend/app"
	"github.com/thienty1207/BWP-Staff/backend/config"
)

func testServerSettings(address string) config.Config {
	return config.Config{
		BackendBindAddress:            address,
		FrontendOrigin:                "http://localhost:5173",
		BackendShutdownTimeoutSeconds: 1,
	}
}

func TestServeGracefullyStopsOnContextCancellation(t *testing.T) {
	server := app.New(app.AppState{}, testServerSettings("127.0.0.1:0"))
	started := make(chan struct{})
	server.Hooks().OnListen(func(fiber.ListenData) error {
		close(started)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		finished <- serve(ctx, server, testServerSettings("127.0.0.1:0"))
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start listening")
	}

	cancel()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatalf("graceful shutdown returned an error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("graceful shutdown did not finish")
	}
}

func TestServeReturnsListenerFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve test listener: %v", err)
	}
	defer listener.Close()

	server := app.New(app.AppState{}, testServerSettings(listener.Addr().String()))
	err = serve(context.Background(), server, testServerSettings(listener.Addr().String()))
	if err == nil || !strings.Contains(err.Error(), "listen") {
		t.Fatalf("expected listener failure, got %v", err)
	}
}
