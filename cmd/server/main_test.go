package main

import (
	"context"
	"errors"
	"testing"

	"github.com/alvaroroco/mcp-server-steam/internal/steam"
	"github.com/mark3labs/mcp-go/mcp"
)

// fakeSteamClient is a test double implementing steam.ClientInterface.
// It returns whatever values are pre-loaded into its fields.
type fakeSteamClient struct {
	recentGamesResp *steam.RecentlyPlayedGamesResponse
	recentGamesErr  error
	ownedGamesResp  *steam.GetOwnedGamesResponse
	ownedGamesErr   error
	// callCount lets tests verify that the real client was (or wasn't) called.
	recentGamesCalls int
	ownedGamesCalls  int
}

func (f *fakeSteamClient) GetRecentGames(_ context.Context, _ string) (*steam.RecentlyPlayedGamesResponse, error) {
	f.recentGamesCalls++
	return f.recentGamesResp, f.recentGamesErr
}

func (f *fakeSteamClient) GetOwnedGames(_ context.Context, _ string) (*steam.GetOwnedGamesResponse, error) {
	f.ownedGamesCalls++
	return f.ownedGamesResp, f.ownedGamesErr
}

// newCallToolRequest is a helper that builds a mcp.CallToolRequest with the
// given "steam_id" argument, mimicking what the MCP framework does at runtime.
func newCallToolRequest(steamID string) mcp.CallToolRequest {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"steam_id": steamID,
	}
	return req
}

// validSteamID is a syntactically valid Steam64 ID used across all happy-path tests.
const validSteamID = "76561197960265728"

// ─────────────────────────────────────────────────────────────────────────────
// makeGetRecentGamesHandler tests
// ─────────────────────────────────────────────────────────────────────────────

// TestGetRecentGamesHandler_InvalidSteamID verifies that the handler returns a
// tool error (IsError == true) and never calls the client when steam_id is invalid.
func TestGetRecentGamesHandler_InvalidSteamID(t *testing.T) {
	fake := &fakeSteamClient{}
	handler := makeGetRecentGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest("abc"))

	if err != nil {
		t.Fatalf("expected nil error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if !result.IsError {
		t.Errorf("expected result.IsError == true for invalid steam_id, got false")
	}
	if fake.recentGamesCalls != 0 {
		t.Errorf("client should NOT be called for invalid steam_id, got %d calls", fake.recentGamesCalls)
	}
}

// TestGetRecentGamesHandler_ValidSteamID verifies that the handler calls the
// steam client when steam_id is valid, and returns a non-error result.
func TestGetRecentGamesHandler_ValidSteamID(t *testing.T) {
	fake := &fakeSteamClient{
		recentGamesResp: &steam.RecentlyPlayedGamesResponse{
			Response: steam.RecentGamesResult{
				TotalCount: 1,
				Games: []steam.RecentGame{
					{AppID: 570, Name: "Dota 2", PlaytimeForever: 100},
				},
			},
		},
	}
	handler := makeGetRecentGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest(validSteamID))

	if err != nil {
		t.Fatalf("expected nil error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if result.IsError {
		t.Errorf("expected result.IsError == false for valid steam_id, got true")
	}
	if fake.recentGamesCalls != 1 {
		t.Errorf("expected exactly 1 client call for valid steam_id, got %d", fake.recentGamesCalls)
	}
}

// TestGetRecentGamesHandler_ClientError verifies that when the steam client
// returns an error the handler surfaces it as a tool error, not a Go error.
func TestGetRecentGamesHandler_ClientError(t *testing.T) {
	fake := &fakeSteamClient{
		recentGamesErr: errors.New("connection refused"),
	}
	handler := makeGetRecentGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest(validSteamID))

	if err != nil {
		t.Fatalf("expected nil Go error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if !result.IsError {
		t.Errorf("expected result.IsError == true when client errors, got false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// makeGetOwnedGamesHandler tests
// ─────────────────────────────────────────────────────────────────────────────

// TestGetOwnedGamesHandler_InvalidSteamID verifies that the handler returns a
// tool error and never calls the client when steam_id is invalid.
func TestGetOwnedGamesHandler_InvalidSteamID(t *testing.T) {
	fake := &fakeSteamClient{}
	handler := makeGetOwnedGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest("abc"))

	if err != nil {
		t.Fatalf("expected nil error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if !result.IsError {
		t.Errorf("expected result.IsError == true for invalid steam_id, got false")
	}
	if fake.ownedGamesCalls != 0 {
		t.Errorf("client should NOT be called for invalid steam_id, got %d calls", fake.ownedGamesCalls)
	}
}

// TestGetOwnedGamesHandler_ValidSteamID verifies that the handler calls the
// steam client when steam_id is valid, and returns a non-error result.
func TestGetOwnedGamesHandler_ValidSteamID(t *testing.T) {
	fake := &fakeSteamClient{
		ownedGamesResp: &steam.GetOwnedGamesResponse{
			Response: steam.OwnedGamesResult{
				GameCount: 1,
				Games: []steam.OwnedGame{
					{AppID: 570, Name: "Dota 2", PlaytimeForever: 500},
				},
			},
		},
	}
	handler := makeGetOwnedGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest(validSteamID))

	if err != nil {
		t.Fatalf("expected nil error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if result.IsError {
		t.Errorf("expected result.IsError == false for valid steam_id, got true")
	}
	if fake.ownedGamesCalls != 1 {
		t.Errorf("expected exactly 1 client call for valid steam_id, got %d", fake.ownedGamesCalls)
	}
}

// TestGetOwnedGamesHandler_ClientError verifies that when the steam client
// returns an error the handler surfaces it as a tool error, not a Go error.
func TestGetOwnedGamesHandler_ClientError(t *testing.T) {
	fake := &fakeSteamClient{
		ownedGamesErr: errors.New("timeout"),
	}
	handler := makeGetOwnedGamesHandler(fake, "")

	result, err := handler(context.Background(), newCallToolRequest(validSteamID))

	if err != nil {
		t.Fatalf("expected nil Go error from handler, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if !result.IsError {
		t.Errorf("expected result.IsError == true when client errors, got false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateSteamID startup scenario (static verification)
// ─────────────────────────────────────────────────────────────────────────────
//
// The log.Fatalf path in main() (invalid STEAM_ID env var → process exits) is
// tested indirectly via TestValidateSteamID in internal/steam/validate_test.go.
// The startup code calls steam.ValidateSteamID and log.Fatalf on non-nil error.
// Since log.Fatalf calls os.Exit(1), it cannot be unit-tested without process
// isolation. The validate package's 13 table-driven tests cover all invalid
// inputs, so the startup guard is proven correct by composition.
