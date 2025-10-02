# mcpcli - MCP Command Line Interface

A powerful command-line tool for interacting with Model Context Protocol (MCP) servers.

## Installation

```bash
go install github.com/jmcarbo/fullmcp/cmd/mcpcli@latest
```

Or build from source:

```bash
cd cmd/mcpcli
go build -o mcpcli
```

## Usage

### Basic Commands

#### Ping Server
Test connectivity to an MCP server:

```bash
mcpcli ping
```

#### Server Information
Display server capabilities and statistics:

```bash
mcpcli info
mcpcli info --json  # Output as JSON
```

### Tools

#### List Tools
Display all available tools:

```bash
mcpcli list-tools
mcpcli list-tools --json      # Output as JSON
mcpcli list-tools --verbose   # Show detailed schemas
```

#### Call Tool
Execute a tool with arguments:

```bash
mcpcli call-tool add --args '{"a":5,"b":3}'
mcpcli call-tool search --args '{"query":"golang"}'
mcpcli call-tool my-tool --json  # Output as JSON
```

### Resources

#### List Resources
Display all available resources:

```bash
mcpcli list-resources
mcpcli list-resources --json  # Output as JSON
```

#### Read Resource
Retrieve resource content:

```bash
mcpcli read-resource config://app
mcpcli read-resource file:///path/to/file
mcpcli read-resource db://schema --json
```

### Prompts

#### List Prompts
Display all available prompts:

```bash
mcpcli list-prompts
mcpcli list-prompts --verbose  # Show arguments
mcpcli list-prompts --json     # Output as JSON
```

#### Get Prompt
Retrieve a prompt with arguments:

```bash
mcpcli get-prompt greeting --args name=Alice
mcpcli get-prompt template --args key1=value1,key2=value2
mcpcli get-prompt my-prompt --json
```

## Global Flags

- `-t, --timeout <seconds>` - Request timeout (default: 30)
- `-v, --verbose` - Enable verbose output
- `--help` - Show help information
- `--version` - Display version

## Examples

### Workflow Example

```bash
# 1. Test connection
mcpcli ping

# 2. Get server info
mcpcli info --verbose

# 3. List available tools
mcpcli list-tools

# 4. Call a specific tool
mcpcli call-tool calculate --args '{"operation":"multiply","a":6,"b":7}'

# 5. Read a resource
mcpcli read-resource config://settings

# 6. Get a prompt
mcpcli get-prompt code-review --args file=main.go
```

### JSON Output for Scripting

```bash
# Get tools as JSON and process with jq
mcpcli list-tools --json | jq '.[] | .name'

# Call tool and extract result
mcpcli call-tool get-user --args '{"id":123}' --json | jq '.result'
```

### Using with Pipes

```bash
# Connect to remote server via SSH
ssh user@host 'mcp-server' | mcpcli list-tools

# Chain commands
mcpcli list-resources | grep config
```

## Transport Support

Currently supports **stdio** transport (standard input/output). The CLI connects to MCP servers that communicate via stdin/stdout.

## Exit Codes

- `0` - Success
- `1` - Error occurred

## Development

### Building

```bash
go build -o mcpcli ./cmd/mcpcli
```

### Testing

```bash
# Start a test server
cd examples/basic-server
go run main.go &

# Test CLI commands
mcpcli ping
mcpcli list-tools
```

## Troubleshooting

### Connection Timeout

If you experience timeouts, increase the timeout value:

```bash
mcpcli list-tools --timeout 60
```

### Verbose Mode

Enable verbose mode for detailed debug information:

```bash
mcpcli list-tools --verbose
```

### JSON Parsing

For reliable parsing in scripts, always use `--json` flag:

```bash
mcpcli info --json | jq
```

## Contributing

Contributions are welcome! Please submit issues and pull requests to the main repository.

## License

See the main project LICENSE file.
