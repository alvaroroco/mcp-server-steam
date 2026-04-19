package main

import (
	"os"
	"testing"
)

// helper: set env var, return a restore function
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("setenv %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	_ = os.Unsetenv(key)
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		}
	})
}

// TestLoadConfig_Defaults verifies that when no env vars are set
// the server defaults to stdio transport on port 8080.
func TestLoadConfig_Defaults(t *testing.T) {
	unsetEnv(t, "MCP_TRANSPORT")
	unsetEnv(t, "MCP_PORT")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for default config, got: %v", err)
	}
	if cfg.Transport != "stdio" {
		t.Errorf("expected default Transport = %q, got %q", "stdio", cfg.Transport)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default Port = %d, got %d", 8080, cfg.Port)
	}
}

// TestLoadConfig_ExplicitStdio verifies that MCP_TRANSPORT=stdio is accepted.
func TestLoadConfig_ExplicitStdio(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "stdio")
	unsetEnv(t, "MCP_PORT")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for stdio transport, got: %v", err)
	}
	if cfg.Transport != "stdio" {
		t.Errorf("expected Transport = %q, got %q", "stdio", cfg.Transport)
	}
}

// TestLoadConfig_ExplicitHTTP verifies that MCP_TRANSPORT=http is accepted.
func TestLoadConfig_ExplicitHTTP(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "http")
	unsetEnv(t, "MCP_PORT")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for http transport, got: %v", err)
	}
	if cfg.Transport != "http" {
		t.Errorf("expected Transport = %q, got %q", "http", cfg.Transport)
	}
}

// TestLoadConfig_CustomPort verifies that MCP_PORT is parsed correctly.
func TestLoadConfig_CustomPort(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "http")
	setEnv(t, "MCP_PORT", "9090")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error for custom port, got: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected Port = %d, got %d", 9090, cfg.Port)
	}
}

// TestLoadConfig_InvalidTransport verifies that an unsupported MCP_TRANSPORT
// value causes LoadConfig to return a non-nil error.
func TestLoadConfig_InvalidTransport(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "websocket")
	unsetEnv(t, "MCP_PORT")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected an error for invalid transport, got nil")
	}
}

// TestLoadConfig_InvalidPort verifies that a non-numeric MCP_PORT value
// causes LoadConfig to return a non-nil error.
func TestLoadConfig_InvalidPort(t *testing.T) {
	setEnv(t, "MCP_TRANSPORT", "http")
	setEnv(t, "MCP_PORT", "not-a-number")

	_, err := LoadConfig()
	if err == nil {
		t.Error("expected an error for invalid port, got nil")
	}
}
