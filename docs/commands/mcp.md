# smoke mcp

Start an MCP (Model Context Protocol) server for Claude Desktop integration.

## Usage

```bash
smoke mcp
```

## Description

Starts a stdio-based MCP server that exposes smoke test operations as tools. Designed for use with Claude Desktop and other MCP-compatible clients.

## Claude Desktop Configuration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "cosmo-smoke": {
      "command": "smoke",
      "args": ["mcp"]
    }
  }
}
```

## Examples

```bash
smoke mcp                              # Start MCP server (stdio transport)
echo '{"jsonrpc":"2.0",...}' | smoke mcp  # Direct JSON-RPC input
```
