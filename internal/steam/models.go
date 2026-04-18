package steam

// RecentlyPlayedGamesResponse wraps the Steam API response for GetRecentlyPlayedGames.
type RecentlyPlayedGamesResponse struct {
	Response RecentGamesResult `json:"response"`
}

// RecentGamesResult holds the total count and game slice for recently played games.
type RecentGamesResult struct {
	TotalCount int          `json:"total_count"`
	Games      []RecentGame `json:"games"`
}

// RecentGame represents a single game entry in the recently played games list.
type RecentGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  int    `json:"playtime_2weeks,omitempty"`
	PlaytimeForever int    `json:"playtime_forever"`
	ImgIconURL      string `json:"img_icon_url,omitempty"`
}

// GetOwnedGamesResponse wraps the Steam API response for GetOwnedGames.
type GetOwnedGamesResponse struct {
	Response OwnedGamesResult `json:"response"`
}

// OwnedGamesResult holds the game count and game slice for the player's library.
type OwnedGamesResult struct {
	GameCount int         `json:"game_count"`
	Games     []OwnedGame `json:"games"`
}

// OwnedGame represents a single game entry in the player's owned games library.
type OwnedGame struct {
	AppID                    int    `json:"appid"`
	Name                     string `json:"name"`
	PlaytimeForever          int    `json:"playtime_forever"`
	ImgIconURL               string `json:"img_icon_url,omitempty"`
	HasCommunityVisibleStats bool   `json:"has_community_visible_stats,omitempty"`
}
