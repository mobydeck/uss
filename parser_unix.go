package uss

import (
	"fmt"
	"strconv"
	"strings"
)

// parseUnixEntry parses a UNIX socket line (u_str, u_dgr, u_seq)
func parseUnixEntry(line string, lineNum int) (Entry, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Entry{}, fmt.Errorf("insufficient fields for UNIX entry (need at least 4, got %d)", len(fields))
	}

	entry := Entry{
		Netid:    fields[0],
		UnixType: fields[0], // Mirror netid to unixType
		State:    fields[1],
	}

	// Parse RecvQ
	recvQ, err := strconv.Atoi(fields[2])
	if err != nil {
		return Entry{}, fmt.Errorf("invalid RecvQ value %q: %w", fields[2], err)
	}
	entry.RecvQ = recvQ

	// Parse SendQ
	sendQ, err := strconv.Atoi(fields[3])
	if err != nil {
		return Entry{}, fmt.Errorf("invalid SendQ value %q: %w", fields[3], err)
	}
	entry.SendQ = sendQ

	// Parse remaining UNIX-specific fields
	// Format typically: Path ID Peer PeerID [Process...]
	// But can be shorter, e.g., just: Path ID Peer PeerID

	fieldIdx := 4

	// Field 5: UnixPath (could be filesystem path, abstract socket @..., or *)
	if fieldIdx < len(fields) {
		entry.UnixPath = fields[fieldIdx]
		fieldIdx++
	}

	// Field 6: UnixID (numeric identifier)
	if fieldIdx < len(fields) {
		entry.UnixID = fields[fieldIdx]
		fieldIdx++
	}

	// Field 7: UnixPeer (often "*")
	if fieldIdx < len(fields) {
		entry.UnixPeer = fields[fieldIdx]
		fieldIdx++
	}

	// Field 8: UnixPeerID (often "0")
	if fieldIdx < len(fields) {
		entry.UnixPeerID = fields[fieldIdx]
		fieldIdx++
	}

	// Remaining fields are Process metadata
	if fieldIdx < len(fields) {
		entry.ProcessRaw = strings.Join(fields[fieldIdx:], " ")
		extractProcessMetadata(&entry, entry.ProcessRaw)
	}

	return entry, nil
}
