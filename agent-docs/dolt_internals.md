# Dolt Internals Reference

**MANDATORY READ BEFORE:**
- Modifying `internal/data/dolt/`
- Working on ticket store implementations
- Debugging database connection issues

---

## internal/data/dolt/ - Beads Database Access

The `dolt` package implements `data.TicketStore` for reading tickets from Beads/Dolt databases.

**Key files:**
- `resolve.go` - Resolves connection settings via `bd dolt show --json` with local fallback
- `metadata.go` - Connection config struct and port helpers
- `credentials.go` - Password resolution (`BEADS_DOLT_PASSWORD`, credentials file)
- `server.go` - MySQL driver for Dolt server connections
- `store.go` - Main `Store` type implementing `TicketStore`
- `schema.go` - Schema verification utilities

**Connection mode:**
Server mode connects to a running Dolt sql-server via MySQL protocol. Blunderbust does not support beads embedded mode.

**Config resolution (beads 0.62+):**

Blunderbust delegates to beads' canonical resolver:

1. **Primary:** `bd dolt show --json` from the project root
2. **Fallback:** layered local resolution matching beads precedence:
   - `BEADS_DOLT_*` environment variables
   - `.beads/dolt-server.port` (or `~/.beads/shared-server/dolt-server.port` in shared-server mode)
   - `.beads/metadata.json` (optional, machine-local)
   - `.beads/config.yaml` `dolt.*` settings

`metadata.json` alone is no longer required. Fresh clones with only `config.yaml` and/or `issues.jsonl` are valid beads projects once bootstrapped.

**Password resolution:**
1. `BEADS_DOLT_PASSWORD` env var
2. `~/.config/beads/credentials` `[host:port]` section (override path via `BEADS_CREDENTIALS_FILE`)

**Usage:**
```go
store, err := dolt.NewStore(ctx, domain.AppOptions{BeadsDir: ".beads"})
if err != nil {
    // Handle with actionable error message
}
defer store.Close()

tickets, err := store.ListTickets(ctx, data.TicketFilter{
    Status: "open",
    Limit: 10,
})
```

**Error handling:** All errors include context. Common patterns:
- Missing `.beads/` → "Is this a beads project? Run 'bd init'"
- Uninitialized workspace → "Run 'bd init' or 'bd bootstrap'"
- Connection failures → Check server running / database corrupted
- Schema failures → "Try running 'bd init' to repair"