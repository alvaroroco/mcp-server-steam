package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.steampowered.com"

// Client is an HTTP client for the Steam Web API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a new Steam API client with the given API key.
func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetRecentGames returns the list of games the player has played in the last 2 weeks.
// Steam returns HTTP 200 with an empty response body for invalid or private profiles,
// so the caller must check TotalCount to detect this case.
func (c *Client) GetRecentGames(ctx context.Context, steamID string) (*RecentlyPlayedGamesResponse, error) {
	endpoint := fmt.Sprintf("%s/IPlayerService/GetRecentlyPlayedGames/v1/", baseURL)

	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("steamid", steamID)
	params.Set("format", "json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API returned status %d", resp.StatusCode)
	}

	var result RecentlyPlayedGamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// GetOwnedGames returns the complete game library for the given Steam ID.
// Includes app info and free-to-play games. Steam returns HTTP 200 with an empty
// response body for invalid or private profiles, so the caller must check GameCount.
func (c *Client) GetOwnedGames(ctx context.Context, steamID string) (*GetOwnedGamesResponse, error) {
	endpoint := fmt.Sprintf("%s/IPlayerService/GetOwnedGames/v1/", baseURL)

	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("steamid", steamID)
	params.Set("include_appinfo", "true")
	params.Set("include_played_free_games", "true")
	params.Set("format", "json")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("steam API returned status %d", resp.StatusCode)
	}

	var result GetOwnedGamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}
