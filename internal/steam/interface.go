package steam

import "context"

// ClientInterface is the contract that any Steam API client must satisfy.
// It is defined here so that cmd/server and tests can depend on the interface
// instead of the concrete *Client, enabling handler tests with a fake.
type ClientInterface interface {
	GetRecentGames(ctx context.Context, steamID string) (*RecentlyPlayedGamesResponse, error)
	GetOwnedGames(ctx context.Context, steamID string) (*GetOwnedGamesResponse, error)
}

// Compile-time assertion: *Client must satisfy ClientInterface.
var _ ClientInterface = (*Client)(nil)
