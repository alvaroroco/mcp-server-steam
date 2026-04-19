# mcp-server-steam

MCP server that exposes Steam Web API data as tools for AI assistants.
Built with Go and mcp-go.

![MCP Server](https://img.shields.io/badge/MCP-server-blue)

---

## Quick start

docker run -i --rm \
-e STEAM_API_KEY=your_api_key \
-e STEAM_ID=your_steam_id \
alvaroroco1/mcp-server-steam:1.0.0

---

## Overview

This MCP server provides Steam data to AI assistants:

- Recently played games (last 2 weeks)
- Owned games library with playtime

---

## Tools

### get_recent_games

Returns games played in the last 2 weeks.

Parameters:

- steam_id (string) optional if STEAM_ID is set

Output:

- list of recently played games

---

### get_owned_games

Returns full Steam library.

Parameters:

- steam_id (string) optional if STEAM_ID is set

Output:

- list of owned games with playtime

---

## Environment variables

STEAM_API_KEY → required → Steam Web API key
STEAM_ID → optional → default Steam ID (64-bit)
MCP_TRANSPORT → optional → transport mode: "stdio" (default) or "http"
MCP_PORT → optional → HTTP listen port (default: 8080, only used when MCP_TRANSPORT=http)

API key: <https://steamcommunity.com/dev/apikey>
Steam ID: <https://steamid.io>

---

## Transports

The server supports two transport modes controlled by the `MCP_TRANSPORT` environment variable.

### stdio (default)

Used by Claude Desktop and other MCP clients that communicate over standard input/output.
No extra configuration needed — this is the default when `MCP_TRANSPORT` is unset.

MCP configuration for Claude Desktop:

{
  "mcpServers": {
    "steam": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "STEAM_API_KEY",
        "-e", "STEAM_ID",
        "alvaroroco1/mcp-server-steam:latest"
      ]
    }
  }
}

### HTTP/SSE

Starts an HTTP server that exposes an SSE endpoint for MCP communication.
Useful for web-based clients or programmatic access.

Run locally:

MCP_TRANSPORT=http MCP_PORT=8080 STEAM_API_KEY=your_key ./mcp-server-steam

Run with Docker:

docker run --rm \
  -e STEAM_API_KEY=your_api_key \
  -e STEAM_ID=your_steam_id \
  -e MCP_TRANSPORT=http \
  -p 8080:8080 \
  alvaroroco1/mcp-server-steam:latest

The SSE endpoint is available at: http://localhost:8080/sse
The message endpoint is available at: http://localhost:8080/message

---

## MCP configuration

{
  "mcpServers": {
    "steam": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "STEAM_API_KEY",
        "-e", "STEAM_ID",
        "alvaroroco1/mcp-server-steam:1.0.0"
      ]
    }
  }
}

---

## Repository contents

- server.json → MCP manifest
- cmd/ → server entrypoint
- internal/steam → Steam API client

---

## Why this exists

Enables AI assistants to:

- analyze gaming habits
- recommend games
- access Steam libraries programmatically
- build gaming automations

---

## Development

go test ./...

---

## Versioning

1.0.0 initial release
1.0.1 bug fixes
1.1.0 new features
2.0.0 breaking changes

---

## Security

Never expose your Steam API key in public repositories.
