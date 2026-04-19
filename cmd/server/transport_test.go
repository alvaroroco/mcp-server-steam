package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// TestStartupLogMessage_Stdio verifies the log message for stdio transport.
func TestStartupLogMessage_Stdio(t *testing.T) {
	cfg := Config{Transport: "stdio", Port: 8080}
	msg := startupLogMessage(cfg)

	if !strings.Contains(msg, "stdio") {
		t.Errorf("expected message to contain %q, got %q", "stdio", msg)
	}
}

// TestStartupLogMessage_HTTP verifies the log message for HTTP transport includes the port.
func TestStartupLogMessage_HTTP(t *testing.T) {
	cfg := Config{Transport: "http", Port: 9000}
	msg := startupLogMessage(cfg)

	if !strings.Contains(msg, "http") && !strings.Contains(msg, "HTTP") {
		t.Errorf("expected message to reference HTTP transport, got %q", msg)
	}
	if !strings.Contains(msg, "9000") {
		t.Errorf("expected message to contain port %q, got %q", "9000", msg)
	}
}

// TestHTTPBaseURL_DefaultPort verifies the base URL for the default port.
func TestHTTPBaseURL_DefaultPort(t *testing.T) {
	url := httpBaseURL(8080)
	expected := "http://localhost:8080"
	if url != expected {
		t.Errorf("expected base URL %q, got %q", expected, url)
	}
}

// TestHTTPBaseURL_CustomPort verifies the base URL for a custom port.
func TestHTTPBaseURL_CustomPort(t *testing.T) {
	url := httpBaseURL(9000)
	expected := "http://localhost:9000"
	if url != expected {
		t.Errorf("expected base URL %q, got %q", expected, url)
	}
}

// freePort asks the OS for an available TCP port so tests never collide.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not get free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// TestRunHTTPServer_GracefulShutdown verifies that cancelling the context
// causes runHTTPServer to shut down the HTTP/SSE server and return within
// a reasonable timeout — proving the graceful-shutdown path works.
func TestRunHTTPServer_GracefulShutdown(t *testing.T) {
	port := freePort(t)

	mcpServer := server.NewMCPServer("test-shutdown", "1.0.0")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- runHTTPServer(ctx, mcpServer, port)
	}()

	// Wait until the server is accepting connections (up to 500 ms).
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/sse", port))
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Trigger shutdown by cancelling the context.
	cancel()

	// The server must return within 2 seconds.
	select {
	case err := <-done:
		// http.ErrServerClosed is expected — any other non-nil error is a real failure.
		if err != nil && !strings.Contains(err.Error(), "Server closed") {
			t.Errorf("runHTTPServer returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runHTTPServer did not shut down within 2 seconds after context cancellation")
	}
}

// TestRunHTTPServer_ServesSSEEndpoint verifies that the server is reachable
// and responds to /sse while it is running — triangulating liveness.
func TestRunHTTPServer_ServesSSEEndpoint(t *testing.T) {
	port := freePort(t)

	mcpServer := server.NewMCPServer("test-sse-alive", "1.0.0")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = runHTTPServer(ctx, mcpServer, port)
	}()

	// Poll until the server is up (max 500 ms).
	var lastErr error
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/sse", port))
		if err == nil {
			resp.Body.Close()
			lastErr = nil
			// 200 OK proves the server is alive and serving SSE.
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected HTTP 200 from /sse, got %d", resp.StatusCode)
			}
			break
		}
		lastErr = err
		time.Sleep(10 * time.Millisecond)
	}

	if lastErr != nil {
		t.Fatalf("SSE endpoint never became reachable: %v", lastErr)
	}
}
