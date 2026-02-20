# Blunderbuss

Launch development harnesses (OpenCode, Claude, etc.) in tmux windows with context from Beads issues.

## Overview

Blunderbuss provides a TUI-driven workflow for:
- Selecting tickets from your Beads/Dolt issue database
- Choosing harness configurations (which tool, model, agent)
- Launching development sessions in organized tmux windows

## Building

Requires Go 1.22 or later.

### Embedded Dolt Mode (Default)

The default mode connects to a local Dolt database stored in `.beads/dolt/`.
**This mode requires CGO** due to the github.com/dolthub/driver dependency.

```bash
# Build with CGO enabled (default)
make build

# Or with go directly (ensure CGO is enabled)
go build -o blunderbuss ./cmd/blunderbuss
```

### Server Mode (No CGO Required)

If you only use server mode connections (remote Dolt sql-server), you can
build without CGO:

```bash
# Build without CGO
CGO_ENABLED=0 go build -o blunderbuss ./cmd/blunderbuss
```

## Beads Database Connection

Blunderbuss reads ticket data from a Beads/Dolt database. The connection mode
is determined by `.beads/metadata.json`:

### Embedded Mode (Local Database)

Default when `dolt_mode` is not set to `server`:

```json
{
  "database": "dolt",
  "backend": "dolt",
  "dolt_database": "beads_bb"
}
```

### Server Mode (Remote Database)

Activated when `dolt_mode: server` or server connection fields are present:

```json
{
  "database": "dolt",
  "backend": "dolt",
  "dolt_mode": "server",
  "dolt_database": "beads_fo",
  "dolt_server_host": "10.11.0.1",
  "dolt_server_port": 13307,
  "dolt_server_user": "mysql-root"
}
```

For server mode with authentication, set the password via environment variable:

```bash
export BEADS_DOLT_PASSWORD="your-password"
./blunderbuss
```

## Running

```bash
# Run with default config
./blunderbuss

# Run with custom config
./blunderbuss --config /path/to/config.yaml

# Dry run (print commands without executing)
./blunderbuss --dry-run

# Debug mode
./blunderbuss --debug
```

## Development

```bash
# Run linter
make lint

# Run tests
make test

# Clean build artifacts
make clean
```

## Configuration

Configuration is loaded from a YAML file (default: `./config.yaml`).
See the example configuration for harness definitions.

## License

GPL-3.0 License - See LICENSE file for details.
