package steam

import (
	"fmt"
	"strconv"
	"strings"
)

// SteamID64Min is the lowest valid Steam64 individual-account ID.
const SteamID64Min uint64 = 76561197960265728

// SteamID64Max is the highest valid Steam64 individual-account ID.
const SteamID64Max uint64 = 76561202255233023

// ValidateSteamID returns nil if id is a valid 17-digit numeric Steam64 ID
// within the individual-account universe range [SteamID64Min, SteamID64Max].
//
// Leading/trailing whitespace is trimmed before all checks; the trimmed value
// is NOT returned — callers continue using their original string.
//
// Error messages are user-facing and surfaced directly by MCP handlers.
func ValidateSteamID(id string) error {
	trimmed := strings.TrimSpace(id)

	if trimmed == "" {
		return fmt.Errorf("steam_id is empty — pass a 17-digit numeric Steam64 ID")
	}

	if len(trimmed) != 17 {
		return fmt.Errorf("steam_id must be exactly 17 digits, got %d", len(trimmed))
	}

	n, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return fmt.Errorf("steam_id must be a numeric 64-bit Steam ID (not SteamID2, SteamID3, or a vanity URL)")
	}

	if n < SteamID64Min || n > SteamID64Max {
		return fmt.Errorf("steam_id is out of the valid Steam64 range (%d–%d)", SteamID64Min, SteamID64Max)
	}

	return nil
}
