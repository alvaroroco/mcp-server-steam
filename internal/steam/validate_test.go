package steam

import (
	"strings"
	"testing"
)

// TestValidateSteamID covers every scenario from the spec using table-driven tests.
// Package-level access (white-box) so we can reference SteamID64Min/SteamID64Max
// in the boundary test cases.
func TestValidateSteamID(t *testing.T) {
	t.Parallel()

	// A known valid Steam64 ID, in the middle of the range.
	const validID = "76561198000000001"

	tests := []struct {
		name          string
		input         string
		wantErr       bool
		wantErrSubstr string
	}{
		// -----------------------------------------------------------------------
		// Happy paths
		// -----------------------------------------------------------------------
		{
			name:    "valid 17-digit Steam64 ID",
			input:   validID,
			wantErr: false,
		},
		{
			name:    "valid ID with leading/trailing spaces (trimmed internally)",
			input:   "  " + validID + "  ",
			wantErr: false,
		},
		{
			name:    "SteamID64Min boundary — must be valid",
			input:   "76561197960265728", // SteamID64Min
			wantErr: false,
		},
		{
			name:    "SteamID64Max boundary — must be valid",
			input:   "76561202255233023", // SteamID64Max
			wantErr: false,
		},

		// -----------------------------------------------------------------------
		// Empty / whitespace
		// -----------------------------------------------------------------------
		{
			name:          "empty string",
			input:         "",
			wantErr:       true,
			wantErrSubstr: "empty",
		},
		{
			name:          "whitespace-only string",
			input:         "   ",
			wantErr:       true,
			wantErrSubstr: "empty",
		},

		// -----------------------------------------------------------------------
		// Wrong length
		// -----------------------------------------------------------------------
		{
			name:          "too short — 3 digits",
			input:         "123",
			wantErr:       true,
			wantErrSubstr: "17 digit",
		},
		{
			name:          "too long — 19 digits",
			input:         "7656119800000000100",
			wantErr:       true,
			wantErrSubstr: "17 digit",
		},

		// -----------------------------------------------------------------------
		// Non-numeric / format errors
		// -----------------------------------------------------------------------
		{
			name:          "non-numeric alphabetic string — wrong length",
			input:         "abc",
			wantErr:       true,
			// Length check fires first because len("abc") != 17.
			wantErrSubstr: "17 digit",
		},
		{
			name:          "SteamID2 format — non-numeric 17-char-like string",
			input:         "STEAM_0:1:1234567",
			wantErr:       true,
			wantErrSubstr: "numeric",
		},
		{
			name:          "vanity URL — short non-numeric",
			input:         "myvanityurl",
			wantErr:       true,
			// Length check fires first; length != 17.
			wantErrSubstr: "17 digit",
		},

		// -----------------------------------------------------------------------
		// Out of range
		// -----------------------------------------------------------------------
		{
			name:          "below SteamID64Min by 1",
			input:         "76561197960265727", // SteamID64Min - 1
			wantErr:       true,
			wantErrSubstr: "range",
		},
		{
			name:          "above SteamID64Max by 1",
			input:         "76561202255233024", // SteamID64Max + 1
			wantErr:       true,
			wantErrSubstr: "range",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range var
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateSteamID(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("ValidateSteamID(%q) = nil, want an error containing %q", tc.input, tc.wantErrSubstr)
				}
				if tc.wantErrSubstr != "" && !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Errorf("ValidateSteamID(%q) error = %q, want it to contain %q", tc.input, err.Error(), tc.wantErrSubstr)
				}
				return
			}

			// wantErr == false
			if err != nil {
				t.Fatalf("ValidateSteamID(%q) = %v, want nil", tc.input, err)
			}
		})
	}
}
