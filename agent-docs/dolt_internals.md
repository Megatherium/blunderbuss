# Dolt Internals Reference

**MANDATORY READ BEFORE:**
- Modifying `internal/data/dolt/`
- Working on ticket store implementations
- Debugging database connection issues

---

## internal/data/dolt/ - Beads Database Access

The `dolt` package implements `data.TicketStore` for reading tickets from Beads/Dolt databases.

**Key files:**
- `metadata.go` - Parses `.beads/metadata.json` to determine connection settings
- `server.go` - MySQL driver for Dolt server connections
- `store.go` - Main `Store` type implementing `TicketStore`
- `schema.go` - Schema verification utilities

**Connection mode:**
Server mode connects to a running Dolt sql-server via MySQL protocol. Activated by `dolt_mode: server` in metadata.json.

**Build:**
Standard build (~13MB) supports server mode only.

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
- Missing metadata.json → "Is this a beads project? Run 'bd init'"
- Missing dolt directory → "The beads database may not be initialized"
- Connection failures → Check server running / database corrupted
- Schema failures → "Try running 'bd init' to repair"
