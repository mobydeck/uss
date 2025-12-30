package uss

import (
	"strconv"
	"strings"
)

// extractProcessMetadata extracts process-related metadata from the process raw string
// and populates the Entry fields
func extractProcessMetadata(entry *Entry, processRaw string) {
	// Extract users:((name,pid=N,fd=M),...)
	extractUsers(entry, processRaw)

	// Extract simple key:value patterns
	extractKeyValue(entry, processRaw, "uid:", func(val string) {
		if uid, err := strconv.Atoi(val); err == nil {
			entry.UID = &uid
		}
	})

	extractKeyValue(entry, processRaw, "ino:", func(val string) {
		if ino, err := strconv.ParseUint(val, 10, 64); err == nil {
			entry.Inode = &ino
		}
	})

	extractKeyValue(entry, processRaw, "cgroup:", func(val string) {
		entry.CGroup = &val
	})

	extractKeyValue(entry, processRaw, "v6only:", func(val string) {
		if v6, err := strconv.Atoi(val); err == nil {
			entry.V6Only = &v6
		}
	})

	extractKeyValue(entry, processRaw, "fwmark:", func(val string) {
		entry.FWMark = &val
	})

	extractKeyValue(entry, processRaw, "sk:", func(val string) {
		entry.Sk = &val
	})

	extractKeyValue(entry, processRaw, "dev:", func(val string) {
		entry.Dev = &val
	})

	// Check for peers: marker
	if strings.Contains(processRaw, "peers:") {
		peers := "peers:"
		entry.Peers = &peers
	}
}

// extractUsers parses users:((name,pid=N,fd=M),...) format
func extractUsers(entry *Entry, processRaw string) {
	// Find users:(( and then manually extract until matching ))
	startMarker := "users:(("
	idx := strings.Index(processRaw, startMarker)
	if idx == -1 {
		return
	}

	// Start after "users:(("
	start := idx + len(startMarker)

	// Find the matching )) by counting parentheses
	parenCount := 2 // We already have ((
	end := start
	for end < len(processRaw) && parenCount > 0 {
		if processRaw[end] == '(' {
			parenCount++
		} else if processRaw[end] == ')' {
			parenCount--
		}
		end++
	}

	if parenCount != 0 {
		// Malformed, couldn't find matching parentheses
		return
	}

	// Extract content between (( and ))
	// end is now pointing after the last ), so we need to go back 2 positions
	usersContent := processRaw[start : end-2]

	// Split by ),( to get individual user entries
	userEntries := strings.Split(usersContent, "),(")

	for _, userEntry := range userEntries {
		user := parseUserEntry(userEntry)
		if user.Name != "" {
			entry.Users = append(entry.Users, user)
		}
	}
}

// parseUserEntry parses a single user entry like: "name",pid=N,fd=M
func parseUserEntry(s string) User {
	user := User{}

	// Split by comma
	parts := splitUserParts(s)

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Name is quoted
		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			user.Name = strings.Trim(part, `"`)
			continue
		}

		// pid=N
		if strings.HasPrefix(part, "pid=") {
			if pid, err := strconv.Atoi(strings.TrimPrefix(part, "pid=")); err == nil {
				user.PID = pid
			}
			continue
		}

		// fd=M
		if strings.HasPrefix(part, "fd=") {
			if fd, err := strconv.Atoi(strings.TrimPrefix(part, "fd=")); err == nil {
				user.FD = fd
			}
			continue
		}
	}

	return user
}

// splitUserParts splits by comma but respects quoted strings
func splitUserParts(s string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for _, ch := range s {
		if ch == '"' {
			inQuote = !inQuote
			current.WriteRune(ch)
		} else if ch == ',' && !inQuote {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// extractKeyValue finds key:value patterns and calls the handler with the value
func extractKeyValue(entry *Entry, processRaw, key string, handler func(string)) {
	idx := strings.Index(processRaw, key)
	if idx == -1 {
		return
	}

	// Start after the key
	start := idx + len(key)
	if start >= len(processRaw) {
		return
	}

	// Extract value until whitespace or end
	end := start
	for end < len(processRaw) && !isWhitespace(processRaw[end]) {
		end++
	}

	value := processRaw[start:end]
	if value != "" {
		handler(value)
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
