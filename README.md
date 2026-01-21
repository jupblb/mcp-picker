# mcp-picker

Interactive TUI for selecting MCP servers to use with [Amp] or [Claude Code].

![demo]

## Usage

``` bash
# Amp (default)
amp --mcp-config $(mcp-picker)

# Claude Code
claude --mcp-config $(mcp-picker -a claude)
```

### Agent Formats

Use `-a`/`--agent` to specify the output format:

| Agent  | Flag        | Output                           |
|--------|-------------|----------------------------------|
| amp    | `-a amp`    | Raw server configs (default)     |
| claude | `-a claude` | Wrapped in `{"mcpServers": ...}` |

## Configuration

Create `~/.config/mcp-picker/servers.json`:

``` json
{
  "github": {
    "command": "nix",
    "args": ["run", "github:NixOS/nixpkgs#github-mcp-server", "--", "stdio"],
    "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}" }
  },
  "linear": {
    "url": "https://mcp.linear.app/mcp"
  }
}
```

## Build

``` bash
go build ./...
# or
nix build
```

  [Amp]: https://ampcode.com
  [Claude Code]: https://claude.com/product/claude-code
  [demo]: https://github.com/user-attachments/assets/35f23ccb-6520-410a-8fc0-b52440de2f16
