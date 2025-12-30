package uss

import (
	"fmt"
	"strconv"
	"strings"
)

// parseInetEntry parses an INET socket line (tcp, udp, etc.)
func parseInetEntry(line string, lineNum int) (Entry, error) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return Entry{}, fmt.Errorf("insufficient fields for INET entry (need at least 6, got %d)", len(fields))
	}

	entry := Entry{
		Netid: fields[0],
		State: fields[1],
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

	// Parse Local address:port
	localToken := fields[4]
	entry.Local = localToken
	localAddr, localPort, localIface := splitAddressPort(localToken)
	entry.LocalAddress = localAddr
	entry.LocalPort = localPort
	if localIface != "" {
		entry.Interface = localIface
	}

	// Parse Peer address:port
	peerToken := fields[5]
	entry.Peer = peerToken
	peerAddr, peerPort, _ := splitAddressPort(peerToken)
	entry.PeerAddress = peerAddr
	entry.PeerPort = peerPort

	// Remaining fields are Process metadata
	if len(fields) > 6 {
		entry.ProcessRaw = strings.Join(fields[6:], " ")
		extractProcessMetadata(&entry, entry.ProcessRaw)
	}

	return entry, nil
}

// splitAddressPort splits an address:port token into components
// Handles: IPv4, IPv6 (bracketed), interface suffixes with %, wildcard *, etc.
func splitAddressPort(token string) (addr, port, iface string) {
	// Handle special cases
	if token == "*" {
		return "*", "*", ""
	}
	if token == ":*" {
		return "", "*", ""
	}

	// Extract interface suffix (e.g., %lo, %incusbr0)
	if idx := strings.Index(token, "%"); idx != -1 {
		// Find the interface name
		remaining := token[idx+1:]
		// Interface name ends at : or end of string
		colonIdx := strings.Index(remaining, ":")
		if colonIdx != -1 {
			iface = remaining[:colonIdx]
			token = token[:idx] + remaining[colonIdx:]
		} else {
			iface = remaining
			token = token[:idx]
		}
	}

	// Handle IPv6 with brackets: [addr]:port or [addr]
	if strings.HasPrefix(token, "[") {
		closeIdx := strings.Index(token, "]")
		if closeIdx != -1 {
			addr = token[:closeIdx+1] // Include brackets
			remainder := token[closeIdx+1:]
			if strings.HasPrefix(remainder, ":") {
				port = remainder[1:]
			}
			return
		}
	}

	// Handle wildcard formats like *:989 or _:8090_
	if strings.HasPrefix(token, "*:") {
		return "*", token[2:], iface
	}
	if strings.HasPrefix(token, "_:") && strings.HasSuffix(token, "_") {
		// Format: _:port_
		port = strings.TrimSuffix(strings.TrimPrefix(token, "_:"), "_")
		return "_", port, iface
	}

	// Standard case: split on last colon for port
	lastColon := strings.LastIndex(token, ":")
	if lastColon != -1 {
		addr = token[:lastColon]
		port = token[lastColon+1:]
		return
	}

	// No port separator found
	addr = token
	return
}
