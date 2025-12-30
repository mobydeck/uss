package uss

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Parse reads ss output from r and returns a slice of Entry structs
func Parse(r io.Reader, opt Options) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip error messages and other non-data lines
		if shouldSkipLine(line) {
			continue
		}

		// Skip header line
		if isHeaderLine(line) {
			continue
		}

		// Parse the entry
		entry, err := parseLine(line, lineNum)
		if err != nil {
			if opt.Strict {
				return nil, fmt.Errorf("line %d: %w", lineNum, err)
			}
			// In non-strict mode, skip malformed lines
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	return entries, nil
}

// isHeaderLine detects the header line (starts with "Netid" or "State")
func isHeaderLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "Netid") || strings.HasPrefix(trimmed, "State")
}

// shouldSkipLine detects error messages and other lines to skip
func shouldSkipLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Skip common error messages
	if strings.HasPrefix(trimmed, "Failed to open") {
		return true
	}
	if strings.HasPrefix(trimmed, "Cannot open") {
		return true
	}
	if strings.HasPrefix(trimmed, "Error:") {
		return true
	}

	return false
}

// parseLine determines the socket type and routes to the appropriate parser
func parseLine(line string, lineNum int) (Entry, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Entry{}, fmt.Errorf("insufficient fields (need at least 4, got %d)", len(fields))
	}

	firstField := fields[0]

	// Check if first field is Netid (tcp, udp, u_str, etc.) or State (LISTEN, UNCONN, etc.)
	// If it's a state, the Netid column is missing (e.g., from ss -tnl without Netid in output)
	if isStateValue(firstField) {
		// Netid column is missing, prepend "tcp" as default and reparse
		line = "tcp " + line
		fields = strings.Fields(line)
	}

	netid := fields[0]

	// Detect UNIX sockets by netid prefix
	if strings.HasPrefix(netid, "u_") {
		return parseUnixEntry(line, lineNum)
	}

	// Otherwise, treat as INET socket
	return parseInetEntry(line, lineNum)
}

// isStateValue checks if a string is a socket state value
func isStateValue(s string) bool {
	states := map[string]bool{
		"LISTEN":     true,
		"UNCONN":     true,
		"ESTAB":      true,
		"SYN-SENT":   true,
		"SYN-RECV":   true,
		"FIN-WAIT1":  true,
		"FIN-WAIT2":  true,
		"TIME-WAIT":  true,
		"CLOSE":      true,
		"CLOSE-WAIT": true,
		"LAST-ACK":   true,
		"CLOSING":    true,
	}
	return states[s]
}
