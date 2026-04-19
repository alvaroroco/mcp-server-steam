package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvaroroco/mcp-server-steam/internal/steam"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		log.Fatal("STEAM_API_KEY is not set")
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	steamClient := steam.New(apiKey)
	defaultSteamID := os.Getenv("STEAM_ID")
	if defaultSteamID != "" {
		if err := steam.ValidateSteamID(defaultSteamID); err != nil {
			log.Fatalf("invalid STEAM_ID env var: %v", err)
		}
	}

	s := server.NewMCPServer(
		"mcp-server-steam",
		"1.0.0",
		server.WithDescription("MCP server exposing Steam Web API tools"),
	)

	steamIDDesc := mcp.Description("The 64-bit Steam ID of the player. Optional if STEAM_ID env var is set.")

	recentGamesTool := mcp.NewTool(
		"get_recent_games",
		mcp.WithDescription("Returns the list of games the Steam user has played in the last 2 weeks"),
		mcp.WithString("steam_id", steamIDDesc),
	)
	s.AddTool(recentGamesTool, makeGetRecentGamesHandler(steamClient, defaultSteamID))

	ownedGamesTool := mcp.NewTool(
		"get_owned_games",
		mcp.WithDescription("Returns the complete game library for the given Steam user"),
		mcp.WithString("steam_id", steamIDDesc),
	)
	s.AddTool(ownedGamesTool, makeGetOwnedGamesHandler(steamClient, defaultSteamID))

	log.Printf(startupLogMessage(cfg))

	switch cfg.Transport {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("server error: %v", err)
		}

	case "http":
		log.Printf("HTTP/SSE server listening on :%d", cfg.Port)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := runHTTPServer(ctx, s, cfg.Port); err != nil {
			log.Fatalf("HTTP/SSE server error: %v", err)
		}
	}
}

// startupLogMessage returns the log message announcing the active transport.
// Extracted as a pure function so it can be unit-tested without side effects.
func startupLogMessage(cfg Config) string {
	switch cfg.Transport {
	case "http":
		return fmt.Sprintf("Starting MCP server with HTTP transport on port %d", cfg.Port)
	default:
		return fmt.Sprintf("Starting MCP server with %s transport", cfg.Transport)
	}
}

// httpBaseURL builds the base URL for the SSE server given a port number.
func httpBaseURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// runHTTPServer starts the SSE HTTP server on the given port and blocks until
// ctx is cancelled, at which point it performs a graceful shutdown.
// It returns nil on clean shutdown and any other error from the server itself.
func runHTTPServer(ctx context.Context, s *server.MCPServer, port int) error {
	sseServer := server.NewSSEServer(s,
		server.WithBaseURL(httpBaseURL(port)),
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- sseServer.Start(fmt.Sprintf(":%d", port))
	}()

	select {
	case <-ctx.Done():
		log.Printf("Shutting down HTTP/SSE server…")
		if err := sseServer.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown error: %v", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func makeGetRecentGamesHandler(client steam.ClientInterface, defaultSteamID string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		steamID := req.GetString("steam_id", defaultSteamID)

		if err := steam.ValidateSteamID(steamID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := client.GetRecentGames(ctx, steamID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("steam API error: %v", err)), nil
		}

		if resp.Response.TotalCount == 0 {
			return mcp.NewToolResultError(
				"no recent games found — check that the steam_id is correct and the profile is not private",
			), nil
		}

		out, err := json.Marshal(resp.Response.Games)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("serialisation error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}
}

func makeGetOwnedGamesHandler(client steam.ClientInterface, defaultSteamID string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		steamID := req.GetString("steam_id", defaultSteamID)

		if err := steam.ValidateSteamID(steamID); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := client.GetOwnedGames(ctx, steamID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("steam API error: %v", err)), nil
		}

		if resp.Response.GameCount == 0 {
			return mcp.NewToolResultError(
				"no owned games found — check that the steam_id is correct and the profile is not private",
			), nil
		}

		out, err := json.Marshal(resp.Response.Games)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("serialisation error: %v", err)), nil
		}

		return mcp.NewToolResultText(string(out)), nil
	}
}
