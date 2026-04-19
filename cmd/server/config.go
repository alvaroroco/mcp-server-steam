package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultTransport = "stdio"
	defaultPort      = 8080
)

// Config holds the runtime configuration for the MCP server.
type Config struct {
	Transport string
	Port      int
}

// LoadConfig reads the MCP_TRANSPORT and MCP_PORT environment variables and
// returns a validated Config. It returns an error if any value is invalid.
func LoadConfig() (Config, error) {
	transport := os.Getenv("MCP_TRANSPORT")
	if transport == "" {
		transport = defaultTransport
	}

	portStr := os.Getenv("MCP_PORT")
	port := defaultPort
	if portStr != "" {
		parsed, err := strconv.Atoi(portStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid MCP_PORT %q: must be a number", portStr)
		}
		port = parsed
	}

	cfg := Config{
		Transport: transport,
		Port:      port,
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validateConfig returns an error if the configuration contains unsupported values.
func validateConfig(c Config) error {
	switch c.Transport {
	case "stdio", "http":
		return nil
	default:
		return fmt.Errorf("unsupported MCP_TRANSPORT %q: must be \"stdio\" or \"http\"", c.Transport)
	}
}
