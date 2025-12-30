# uss - Unix Socket Statistics Parser

`uss` is a Go library and CLI tool for parsing Linux socket statistics from the `ss` command output. It supports both INET sockets (tcp, udp) and UNIX sockets (u_str, u_dgr, u_seq) with multiple output formats.

## Features

- **Parse `ss` output**: Handles both INET and UNIX socket formats
- **Multiple output formats**: Human-readable tables, CSV, JSON, and YAML
- **Robust parsing**: Handles IPv4, IPv6 (with brackets and scopes), interface suffixes, abstract sockets
- **Process metadata extraction**: Extracts users, UID, inode, cgroup, and more
- **Reusable library**: Clean API for use in other Go applications
- **CLI tool**: Built with Cobra for easy command-line use

## Installation

### From Source

```bash
go install github.com/mobydeck/uss/cmd/uss@latest
```

### Build Locally

```bash
git clone https://github.com/mobydeck/uss
cd uss
go build -o uss ./cmd/uss
```

## CLI Usage

### Basic Usage

By default, `uss` reads from stdin and outputs human-readable format:

```bash
ss -tunap | uss
```

### Executing ss Directly

Use the `--ss` flag to execute the `ss` command directly with custom flags:

```bash
# Listen on all TCP ports
uss --ss tnl

# Show all TCP and UDP sockets with process information
uss --ss tunap

# Get UNIX sockets with all details
uss --ss peanutlx

# With leading dash (also works)
uss --ss -tnl
```

The `--ss` flag automatically parses output even when the Netid column is missing (e.g., from `ss -tnl`), inferring it as `tcp` by default. The leading dash in ss flags is optional.

### Output Formats

**Human-readable (default):**
```bash
ss -tunap | uss
```

**JSON:**
```bash
ss -tunap | uss --json
```

**Pretty JSON:**
```bash
ss -tunap | uss --json --pretty
```

**YAML:**
```bash
ss -peanutlx | uss --yaml
```

**CSV:**
```bash
ss -tunap | uss --csv > sockets.csv
```

### Reading from File

You can also read from a file instead of stdin:

```bash
ss -tunap > sockets.txt
uss --json sockets.txt
```

### Raw Fields

By default, `uss` hides raw/unparsed fields for cleaner output. These include:
- `processRaw` - The unparsed process metadata string
- `local` - The raw local address:port token (before parsing)
- `peer` - The raw peer address:port token (before parsing)

To include these raw fields in the output, use the `-r` or `--raw` flag:

```bash
# Include raw fields in JSON output
ss -tunap | uss --json --raw

# Include raw fields in CSV output
ss -tunap | uss --csv --raw

# Include raw fields in YAML output
ss -tunap | uss --yaml --raw

# Include raw fields in human-readable output
ss -tunap | uss --raw
```

This is useful when you need to see the original unparsed data or for debugging parsing issues.

### Filtering Results

Filter socket entries using the `-q` or `--query` flag:

#### Query Syntax

- Format: `field=value1,value2[,valueN]`
- Multiple values per field: comma-separated (OR condition)
- Multiple fields: space or semicolon separated (AND condition)

#### Examples

**Single field:**
```bash
# Find all TCP sockets
ss -tunap | uss -q "netid=tcp"

# Find sockets on port 22 or 80
ss -tunap | uss -q "port=22,80"
```

**Multiple fields (AND):**
```bash
# TCP sockets on port 22 or 80 with uid 0 or 1000
ss -tunap | uss -q "netid=tcp port=22,80 uid=0,1000"

# Using semicolon separator
ss -tunap | uss -q "netid=tcp;port=22,80;uid=0,1000"
```

**Wildcards:**
```bash
# All LISTEN sockets
ss -tunap | uss -q "state=LISTEN*"

# Unix sockets in /run directory
ss -peanutlx | uss -q "unixpath=*/run/*"
```

**Ranges:**
```bash
# Ports between 1000 and 2000
ss -tunap | uss -q "port=1000-2000"

# UID between 1000 and 2000
ss -tunap | uss -q "uid=1000-2000"
```

**Substring matching:**
```bash
# Find sockets with "docker" in path
ss -peanutlx | uss -q "unixpath=docker"

# Find sockets with specific cgroup
ss -tunap | uss -q "cgroup=systemd"
```

**Using --ss flag with output formats:**
```bash
# Execute ss directly and output as JSON
uss --ss tunap --json

# Execute ss directly and output as YAML with filtering
uss --ss peanutlx --yaml -q unixpath=/run/*

# Execute ss and export as CSV
uss --ss tunap --csv > all_sockets.csv

# Execute ss and output as pretty JSON
uss --ss tnl --json --pretty
```

**Combined queries with output format:**
```bash
# TCP LISTEN sockets on ports 22 or 80 with uid 0, output as JSON
ss -tunap | uss -q "netid=tcp state=LISTEN port=22,80 uid=0" --json

# Filter and output as CSV
ss -tunap | uss -q "netid=tcp" --csv > tcp_sockets.csv
```

#### Supported Fields

**Common:**
- `netid` - Socket type (tcp, udp, u_str, etc.)
- `state` - Socket state (LISTEN, ESTAB, UNCONN, etc.)
- `recvq` - Receive queue size
- `sendq` - Send queue size

**INET-specific:**
- `port` - Local or peer port (convenience field, matches both)
- `address` - Local or peer address (convenience field, matches both)
- `localport` - Local port
- `peerport` - Peer port
- `localaddress` - Local address
- `peeraddress` - Peer address
- `interface` - Network interface

**UNIX-specific:**
- `unixtype` - Unix socket type (u_str, u_dgr, u_seq)
- `unixpath` - Unix socket path (supports wildcards/substring)
- `unixid` - Unix socket ID
- `unixpeer` - Unix peer socket
- `unixpeerid` - Unix peer socket ID

**Metadata:**
- `uid` - User ID (numeric)
- `inode` - Inode number (numeric)
- `cgroup` - Control group path
- `v6only` - IPv6-only flag (0 or 1)
- `fwmark` - Firewall mark
- `sk` - Socket kernel identifier
- `dev` - Device identifier

#### Matching Rules

- **Exact match**: Default for most values (`port=22`)
- **Wildcard**: Use `*` (any sequence) or `?` (single char) (`state=LISTEN*`)
- **Range**: Use `min-max` for numeric fields (`port=1000-2000`)
- **Substring**: For string fields, partial matches work (`unixpath=docker`)
- **Case-insensitive**: All string matching is case-insensitive (`state=listen` matches `LISTEN`)

### Examples

**Parse INET sockets:**
```bash
ss -tunap | uss
```

**Parse UNIX sockets:**
```bash
ss -peanutlx | uss
```

**Generate CSV report:**
```bash
ss -tunap | uss --csv > sockets.csv
```

### Using jq for Advanced Filtering

You can combine `uss` with `jq` for powerful JSON-based filtering and querying:

**Filter by port:**
```bash
# Find all sockets listening on port 80
ss -tunap | uss --json | jq '.[] | select(.localPort == "80")'
```

**Filter by state and protocol:**
```bash
# Find all LISTEN TCP sockets
ss -tunap | uss --json | jq '.[] | select(.state == "LISTEN" and .netid == "tcp")'
```

**Filter by UID and extract specific fields:**
```bash
# Find sockets owned by user 1000, show only address and port
ss -tunap | uss --json | jq '.[] | select(.uid == 1000) | {addr: .localAddress, port: .localPort, state: .state}'
```

**Count sockets by state:**
```bash
# How many sockets in each state
ss -tunap | uss --json | jq 'group_by(.state) | map({state: .[0].state, count: length})'
```

**Find listening ports:**
```bash
# Get all listening ports
ss -tunap | uss --json | jq -r '.[] | select(.state == "LISTEN") | "\(.localAddress):\(.localPort)"'
```

**Filter by cgroup:**
```bash
# Find all sockets in systemd cgroups
ss -tunap | uss --json | jq '.[] | select(.cgroup | startswith("/system.slice"))'
```

**Complex filtering:**
```bash
# TCP sockets that are either LISTEN or ESTAB, and not owned by root
ss -tunap | uss --json | jq '.[] | select((.state == "LISTEN" or .state == "ESTAB") and (.uid // 0) != 0)'
```

**Extract process information:**
```bash
# Get process names and PIDs for each socket
ss -tunap | uss --json | jq '.[] | select(.users | length > 0) | {port: .localPort, processes: .users[].name}'
```

**Format as table:**
```bash
# Create a custom table output
ss -tunap | uss --json | jq -r '.[] | [.netid, .state, .localAddress, .localPort, (.users[0].name // "unknown")] | @tsv'
```

**Find sockets by pattern:**
```bash
# Find all UNIX sockets in /run directory
ss -peanutlx | uss --json | jq '.[] | select(.unixPath | startswith("/run"))'

# Find abstract sockets
ss -peanutlx | uss --json | jq '.[] | select(.unixPath | startswith("@"))'
```

**Export filtered results:**
```bash
# Export only LISTEN TCP sockets to CSV-like format
ss -tunap | uss --json | jq -r '.[] | select(.state == "LISTEN" and .netid == "tcp") | [.netid, .state, .localAddress, .localPort, .uid] | @csv'
```

## Library Usage

### Import

```go
import "github.com/mobydeck/uss"
```

### Basic Example

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/mobydeck/uss"
)

func main() {
    // Sample ss output
    input := `Netid  State    Recv-Q   Send-Q   Local Address:Port   Peer Address:Port
tcp    LISTEN   0        128      0.0.0.0:22              0.0.0.0:*
udp    UNCONN   0        0        127.0.0.53%lo:53        0.0.0.0:*
`

    // Parse the input
    entries, err := uss.Parse(strings.NewReader(input), uss.Options{Strict: false})
    if err != nil {
        fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
        return
    }

    // Print results
    for _, entry := range entries {
        fmt.Printf("%s %s %s:%s\n", 
            entry.Netid, entry.State, entry.LocalAddress, entry.LocalPort)
    }
}
```

### Rendering Output

```go
package main

import (
    "os"
    "strings"

    "github.com/mobydeck/uss"
)

func main() {
    input := `...` // ss output

    entries, _ := uss.Parse(strings.NewReader(input), uss.Options{})

    // Render as pretty JSON (without raw fields)
    uss.RenderJSON(os.Stdout, entries, true, false)

    // Or render as JSON with raw fields
    // uss.RenderJSON(os.Stdout, entries, true, true)

    // Or render as YAML
    // uss.RenderYAML(os.Stdout, entries, false)

    // Or render as CSV
    // uss.RenderCSV(os.Stdout, entries, false)

    // Or render as human-readable
    // uss.RenderHuman(os.Stdout, entries, false)
}
```

### Filtering Results

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/mobydeck/uss"
)

func main() {
    input := `...` // ss output

    entries, _ := uss.Parse(strings.NewReader(input), uss.Options{})

    // Filter for TCP sockets on port 22 or 80
    filtered, err := uss.Filter(entries, "netid=tcp port=22,80")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Filter error: %v\n", err)
        return
    }

    // Render filtered results as JSON
    uss.RenderJSON(os.Stdout, filtered, true, false)
}
```

### Filtering with Conditions Slice

For programmatic filtering where you build conditions dynamically, use `FilterByConditions`:

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/mobydeck/uss"
)

func main() {
    input := `...` // ss output

    entries, _ := uss.Parse(strings.NewReader(input), uss.Options{})

    // Build conditions dynamically
    conditions := []string{
        "netid=tcp",
        "state=LISTEN",
        "port=22,80,443",
    }

    // Filter using conditions slice
    filtered, err := uss.FilterByConditions(entries, conditions)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Filter error: %v\n", err)
        return
    }

    // Use filtered results
    for _, entry := range filtered {
        fmt.Printf("%s:%s (%s)\n", entry.LocalAddress, entry.LocalPort, entry.State)
    }
}
```

## Data Structures

### Entry

The `Entry` struct represents a single socket entry:

```go
type Entry struct {
    // Common fields
    Netid      string
    State      string
    RecvQ      int
    SendQ      int
    ProcessRaw string

    // Extracted metadata
    Users  []User
    UID    *int
    Inode  *uint64
    CGroup *string
    V6Only *int
    FWMark *string
    Sk     *string
    Dev    *string
    Peers  *string

    // INET-specific
    LocalAddress string
    LocalPort    string
    PeerAddress  string
    PeerPort     string
    Local        string
    Peer         string
    Interface    string

    // UNIX-specific
    UnixType   string
    UnixPath   string
    UnixID     string
    UnixPeer   string
    UnixPeerID string
}
```

### User

The `User` struct represents a process using a socket:

```go
type User struct {
    Name string
    PID  int
    FD   int
}
```

### Options

Control parsing behavior:

```go
type Options struct {
    Strict bool // if true, fail on malformed lines; if false, skip and continue
}
```

## Supported Formats

### INET Sockets

- **IPv4**: `0.0.0.0:111`, `127.0.0.1:8080`
- **IPv4 with interface**: `127.0.0.53%lo:53`, `0.0.0.0%incusbr0:67`
- **IPv6**: `[::]:22`, `[fd42:2eff:9d69:efec::1]:53`
- **IPv6 with interface**: `[fe80::1266:6aff:fecf:96d]%incusbr0:53`
- **Wildcard**: `*:989`, `*:8090`
- **Special formats**: `_:8090_`, `:*`

### UNIX Sockets

- **Filesystem paths**: `/run/systemd/private`, `/var/run/docker.sock`
- **Abstract sockets**: `@/var/lib/incus/containers/satis/command`, `@5cbfc`
- **Socket types**: `u_str` (stream), `u_dgr` (datagram), `u_seq` (seqpacket)

### Process Metadata

Extracts from the process field:
- `users:(("name",pid=N,fd=M),...)` → User list
- `uid:<num>` → UID
- `ino:<num>` → Inode
- `cgroup:<path>` → CGroup
- `v6only:<0|1>` → V6Only flag
- `fwmark:<value>` → Firewall mark
- `sk:<value>` → Socket kernel identifier
- `dev:<maj/min>` → Device
- `peers:` → Peers marker

## Testing

Run the test suite:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Author

Built for parsing Linux socket statistics efficiently in Go.
