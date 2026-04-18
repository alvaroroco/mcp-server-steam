# mcp-server-steam

MCP server that exposes Steam Web API data as tools for AI assistants. Built with Go and [mcp-go](https://github.com/mark3labs/mcp-go).

## Tools

### `get_recent_games`

Returns games played in the last 2 weeks.

| Parameter | Type | Required |
|-----------|------|----------|
| `steam_id` | string | Only if `STEAM_ID` env var is not set |

### `get_owned_games`

Returns the complete game library with playtime.

| Parameter | Type | Required |
|-----------|------|----------|
| `steam_id` | string | Only if `STEAM_ID` env var is not set |

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `STEAM_API_KEY` | Yes | Steam Web API key |
| `STEAM_ID` | No | Default Steam ID (64-bit). If set, `steam_id` parameter becomes optional in all tools. |

Get your API key at [steamcommunity.com/dev/apikey](https://steamcommunity.com/dev/apikey).
Find your 64-bit Steam ID at [steamid.io](https://steamid.io).

## Installation

### Option A — Build with Go

```bash
go mod tidy
go build -o mcp-server-steam ./cmd/server
```

### Option B — Build with Docker (no Go required)

```bash
docker build -t mcp-server-steam .
docker create --name tmp mcp-server-steam
docker cp tmp:/app/mcp-server-steam ./mcp-server-steam
docker rm tmp
```

### Add to Claude Code

```bash
claude mcp add mcp-server-steam \
  -e STEAM_API_KEY=your_api_key \
  -e STEAM_ID=your_steam_id \
  -- /path/to/mcp-server-steam
```

## Development

```bash
go test ./...
```
