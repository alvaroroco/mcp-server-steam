package steam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient creates a steam Client that points its HTTP requests to the
// provided base URL instead of the real Steam API.
func newTestClient(apiKey, baseURLOverride string) *Client {
	c := New(apiKey)
	// We override the httpClient transport so requests go to our test server.
	// Because baseURL is a package-level const we cannot override it directly,
	// so we rely on the fact that the Client builds the URL from baseURL —
	// instead we swap the httpClient's transport to a round-tripper that
	// rewrites the host to our test server.
	c.httpClient = &http.Client{
		Transport: &hostRewriteTransport{
			base:      baseURLOverride,
			delegateTo: http.DefaultTransport,
		},
	}
	return c
}

// hostRewriteTransport rewrites the request URL so that the scheme+host
// match the test server while keeping the path and query unchanged.
type hostRewriteTransport struct {
	base       string // e.g. "http://127.0.0.1:PORT"
	delegateTo http.RoundTripper
}

func (t *hostRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request so we don't mutate the original.
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = "http"
	// Strip the scheme from t.base if present and extract just the host.
	base := strings.TrimPrefix(t.base, "http://")
	base = strings.TrimPrefix(base, "https://")
	cloned.URL.Host = base
	return t.delegateTo.RoundTrip(cloned)
}

// ---------------------------------------------------------------------------
// GetRecentGames tests
// ---------------------------------------------------------------------------

// 5.1 — Happy path: server returns two games.
func TestGetRecentGames_HappyPath(t *testing.T) {
	payload := map[string]any{
		"response": map[string]any{
			"total_count": 2,
			"games": []map[string]any{
				{
					"appid":            730,
					"name":             "CS2",
					"playtime_forever": 1200,
					"playtime_2weeks":  60,
					"img_icon_url":     "abc",
				},
				{
					"appid":            570,
					"name":             "Dota 2",
					"playtime_forever": 3000,
					"playtime_2weeks":  120,
					"img_icon_url":     "def",
				},
			},
		},
	}

	srv := httptest.NewServer(jsonHandler(payload, http.StatusOK))
	defer srv.Close()

	client := newTestClient("valid-key", srv.URL)
	result, err := client.GetRecentGames(context.Background(), "76561198000000001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Response.Games) != 2 {
		t.Fatalf("expected 2 games, got %d", len(result.Response.Games))
	}
	first := result.Response.Games[0]
	if first.AppID != 730 {
		t.Errorf("expected AppID 730, got %d", first.AppID)
	}
	if first.Name != "CS2" {
		t.Errorf("expected Name CS2, got %q", first.Name)
	}
	if result.Response.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", result.Response.TotalCount)
	}
}

// 5.2 — Empty response (private/invalid profile): Steam returns 200 with an
// empty response object. The client returns a non-nil result with TotalCount==0
// and an empty Games slice — it does NOT error. The caller (handler) is
// responsible for detecting this. This test documents that contract.
func TestGetRecentGames_EmptyResponse_PrivateProfile(t *testing.T) {
	payload := map[string]any{
		"response": map[string]any{},
	}

	srv := httptest.NewServer(jsonHandler(payload, http.StatusOK))
	defer srv.Close()

	client := newTestClient("valid-key", srv.URL)
	result, err := client.GetRecentGames(context.Background(), "76561198000000001")

	// The client must NOT error on an empty body — it returns the struct as-is.
	if err != nil {
		t.Fatalf("client should not error on empty response, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even for private profiles")
	}
	// TotalCount defaults to zero when the key is absent from the JSON.
	if result.Response.TotalCount != 0 {
		t.Errorf("expected TotalCount 0 for private profile, got %d", result.Response.TotalCount)
	}
	if len(result.Response.Games) != 0 {
		t.Errorf("expected empty Games slice, got %d games", len(result.Response.Games))
	}
}

// 5.3 — Non-2xx status: server returns 403, client must return an error.
func TestGetRecentGames_Non2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := newTestClient("bad-key", srv.URL)
	result, err := client.GetRecentGames(context.Background(), "76561198000000001")

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error to mention 403, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetOwnedGames tests
// ---------------------------------------------------------------------------

// 5.4 — Happy path: server returns a library with games.
func TestGetOwnedGames_HappyPath(t *testing.T) {
	payload := map[string]any{
		"response": map[string]any{
			"game_count": 2,
			"games": []map[string]any{
				{
					"appid":                       730,
					"name":                        "CS2",
					"playtime_forever":             1200,
					"img_icon_url":                "abc",
					"has_community_visible_stats":  true,
				},
				{
					"appid":            570,
					"name":             "Dota 2",
					"playtime_forever": 3000,
					"img_icon_url":     "def",
				},
			},
		},
	}

	srv := httptest.NewServer(jsonHandler(payload, http.StatusOK))
	defer srv.Close()

	client := newTestClient("valid-key", srv.URL)
	result, err := client.GetOwnedGames(context.Background(), "76561198000000001")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Response.GameCount != 2 {
		t.Errorf("expected GameCount 2, got %d", result.Response.GameCount)
	}
	if len(result.Response.Games) == 0 {
		t.Fatal("expected non-empty Games slice")
	}
	first := result.Response.Games[0]
	if first.AppID != 730 {
		t.Errorf("expected first game AppID 730, got %d", first.AppID)
	}
	if !first.HasCommunityVisibleStats {
		t.Errorf("expected HasCommunityVisibleStats true for first game")
	}
}

// 5.5 — Empty library (private/invalid profile): Steam returns 200 with an
// empty response object. Same contract as GetRecentGames — the client returns
// the struct without error; GameCount and Games will be zero/empty.
func TestGetOwnedGames_EmptyLibrary_PrivateProfile(t *testing.T) {
	payload := map[string]any{
		"response": map[string]any{},
	}

	srv := httptest.NewServer(jsonHandler(payload, http.StatusOK))
	defer srv.Close()

	client := newTestClient("valid-key", srv.URL)
	result, err := client.GetOwnedGames(context.Background(), "76561198000000001")

	if err != nil {
		t.Fatalf("client should not error on empty response, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even for private profiles")
	}
	if result.Response.GameCount != 0 {
		t.Errorf("expected GameCount 0 for private profile, got %d", result.Response.GameCount)
	}
	if len(result.Response.Games) != 0 {
		t.Errorf("expected empty Games slice, got %d games", len(result.Response.Games))
	}
}

// 5.6 — Missing/empty STEAM_API_KEY: a client created with an empty key
// sends the request with key="". Our mock verifies the key param is empty and
// returns 403, which the client must surface as an error.
func TestGetRecentGames_EmptyAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// If somehow a key slips through, still reject to keep the test strict.
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	// Construct a client with empty apiKey.
	client := newTestClient("", srv.URL)
	result, err := client.GetRecentGames(context.Background(), "76561198000000001")

	if err == nil {
		t.Fatal("expected error when API key is empty and server returns non-2xx")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// jsonHandler returns an http.Handler that writes the given payload as JSON
// with the given status code.
func jsonHandler(payload any, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(payload)
	})
}
