# mcp-picker

Interactive TUI for selecting MCP servers to use with [Amp].

## Usage

``` bash
# Amp (default)
amp --mcp-config $(mcp-picker)

# Claude Code
claude --mcp-config $(mcp-picker -a claude)
```

### Agent Formats

Use `-a`/`--agent` to specify the output format:

| Agent   | Flag          | Output                          |
|---------|---------------|---------------------------------|
| amp     | `-a amp`      | Raw server configs (default)    |
| claude  | `-a claude`   | Wrapped in `{"mcpServers": ...}`|

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
nix build
```

## Controls

- `j/k` or arrows: navigate
- `space`: toggle selection
- `enter`: confirm
- `q/esc`: cancel

  [Amp]: https://ampcode.com
