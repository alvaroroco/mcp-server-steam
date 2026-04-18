package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/alvaroroco/mcp-server-steam/internal/steam"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		log.Fatal("STEAM_API_KEY is not set")
	}

	steamClient := steam.New(apiKey)
	defaultSteamID := os.Getenv("STEAM_ID")

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

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func makeGetRecentGamesHandler(client *steam.Client, defaultSteamID string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		steamID := req.GetString("steam_id", defaultSteamID)

		if steamID == "" {
			return mcp.NewToolResultError("steam_id is required — pass it as a parameter or set the STEAM_ID env var"), nil
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

func makeGetOwnedGamesHandler(client *steam.Client, defaultSteamID string) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		steamID := req.GetString("steam_id", defaultSteamID)

		if steamID == "" {
			return mcp.NewToolResultError("steam_id is required — pass it as a parameter or set the STEAM_ID env var"), nil
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
