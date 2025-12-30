package uss

import (
	"regexp"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "single field single value",
			query:     "netid=tcp",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "single field multiple values",
			query:     "port=22,80,443",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "multiple fields space separator",
			query:     "netid=tcp port=22,80",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "multiple fields semicolon separator",
			query:     "netid=tcp;port=22,80",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "multiple fields mixed separators",
			query:     "netid=tcp; port=22,80 uid=0",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "invalid syntax no equals",
			query:   "netid:tcp",
			wantErr: true,
		},
		{
			name:      "empty query",
			query:     "",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(result.Conditions) != tt.wantCount {
				t.Errorf("parseQuery() got %d conditions, want %d", len(result.Conditions), tt.wantCount)
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		wantType MatchType
	}{
		{
			name:     "exact match",
			value:    "22",
			wantType: MatchExact,
		},
		{
			name:     "wildcard with asterisk",
			value:    "LISTEN*",
			wantType: MatchWildcard,
		},
		{
			name:     "wildcard with question mark",
			value:    "????.sock",
			wantType: MatchWildcard,
		},
		{
			name:     "range",
			value:    "1000-2000",
			wantType: MatchRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use numeric field type for range test
			fieldType := "string"
			if tt.wantType == MatchRange {
				fieldType = "numeric"
			}

			matcher, err := parseValue(tt.value, fieldType)
			if err != nil {
				t.Fatalf("parseValue() error: %v", err)
			}
			if matcher.Type != tt.wantType {
				t.Errorf("parseValue() got type %d, want %d", matcher.Type, tt.wantType)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	uid0 := 0
	uid1000 := 1000
	ino1 := uint64(12345)
	ino2 := uint64(67890)

	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN",
			LocalAddress: "0.0.0.0", LocalPort: "22",
			PeerAddress: "0.0.0.0", PeerPort: "*",
			UID: &uid0, Inode: &ino1,
		},
		{
			Netid: "tcp", State: "ESTAB",
			LocalAddress: "127.0.0.1", LocalPort: "80",
			PeerAddress: "192.168.1.1", PeerPort: "54321",
			UID: &uid1000, Inode: &ino2,
		},
		{
			Netid: "udp", State: "UNCONN",
			LocalAddress: "0.0.0.0", LocalPort: "53",
			PeerAddress: "0.0.0.0", PeerPort: "*",
			UID: &uid0, Inode: &ino1,
		},
		{
			Netid: "u_str", State: "LISTEN", UnixPath: "/run/docker.sock",
			RecvQ: 0, SendQ: 4096,
			UID: &uid1000, Inode: &ino2,
		},
		{
			Netid: "u_str", State: "LISTEN", UnixPath: "/var/run/test.sock",
			RecvQ: 0, SendQ: 4096,
			UID: &uid0, Inode: &ino1,
		},
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "filter by netid exact match",
			query:     "netid=tcp",
			wantCount: 2,
		},
		{
			name:      "filter by port single value",
			query:     "port=22",
			wantCount: 1,
		},
		{
			name:      "filter by port multiple values",
			query:     "port=22,80",
			wantCount: 2,
		},
		{
			name:      "filter by state",
			query:     "state=listen",
			wantCount: 3,
		},
		{
			name:      "filter by uid",
			query:     "uid=0",
			wantCount: 3,
		},
		{
			name:      "filter by multiple fields AND",
			query:     "netid=tcp uid=0",
			wantCount: 1,
		},
		{
			name:      "filter by multiple fields with multiple values",
			query:     "netid=tcp,udp uid=0",
			wantCount: 2,
		},
		{
			name:      "filter by wildcard pattern",
			query:     "unixpath=*/docker*",
			wantCount: 1,
		},
		{
			name:      "filter by substring",
			query:     "unixpath=docker",
			wantCount: 1,
		},
		{
			name:      "filter by port range",
			query:     "port=1000-60000",
			wantCount: 1, // Only 54321 falls in range
		},
		{
			name:      "filter by inode numeric",
			query:     "inode=12345",
			wantCount: 3,
		},
		{
			name:      "no matches",
			query:     "port=9999",
			wantCount: 0,
		},
		{
			name:      "empty query returns all",
			query:     "",
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Filter(entries, tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(result) != tt.wantCount {
				t.Errorf("Filter() got %d results, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestMatchValue(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		matcher   ValueMatcher
		fieldType string
		wantMatch bool
	}{
		{
			name:  "exact match string",
			value: "LISTEN",
			matcher: ValueMatcher{
				Type:  MatchExact,
				Value: "listen",
			},
			fieldType: "string",
			wantMatch: true,
		},
		{
			name:  "exact match int",
			value: 22,
			matcher: ValueMatcher{
				Type:  MatchExact,
				Value: "22",
			},
			fieldType: "numeric",
			wantMatch: true,
		},
		{
			name:  "wildcard match",
			value: "LISTENING",
			matcher: ValueMatcher{
				Type:  MatchWildcard,
				Value: "LISTEN*",
			},
			fieldType: "string",
			wantMatch: true,
		},
		{
			name:  "wildcard no match",
			value: "ESTAB",
			matcher: ValueMatcher{
				Type:  MatchWildcard,
				Value: "LISTEN*",
			},
			fieldType: "string",
			wantMatch: false,
		},
		{
			name:  "substring match",
			value: "/run/docker.sock",
			matcher: ValueMatcher{
				Type:  MatchExact,
				Value: "docker",
			},
			fieldType: "string",
			wantMatch: true,
		},
		{
			name:  "range match in range",
			value: 1500,
			matcher: ValueMatcher{
				Type:  MatchRange,
				Value: "1000-2000",
				Min:   intPtr(1000),
				Max:   intPtr(2000),
			},
			fieldType: "numeric",
			wantMatch: true,
		},
		{
			name:  "range match out of range",
			value: 3000,
			matcher: ValueMatcher{
				Type:  MatchRange,
				Value: "1000-2000",
				Min:   intPtr(1000),
				Max:   intPtr(2000),
			},
			fieldType: "numeric",
			wantMatch: false,
		},
		{
			name:  "case insensitive",
			value: "TCP",
			matcher: ValueMatcher{
				Type:  MatchExact,
				Value: "tcp",
			},
			fieldType: "string",
			wantMatch: true,
		},
		{
			name:  "numeric exact no substring",
			value: 0,
			matcher: ValueMatcher{
				Type:  MatchExact,
				Value: "0",
			},
			fieldType: "numeric",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchValue(tt.value, tt.matcher, tt.fieldType)
			if result != tt.wantMatch {
				t.Errorf("matchValue() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		pattern   string
		test      string
		wantMatch bool
	}{
		{
			pattern:   "LISTEN*",
			test:      "LISTENING",
			wantMatch: true,
		},
		{
			pattern:   "*/run/*",
			test:      "/var/run/docker.sock",
			wantMatch: true,
		},
		{
			pattern:   "????.sock",
			test:      "test.sock",
			wantMatch: true,
		},
		{
			pattern:   "LISTEN*",
			test:      "ESTAB",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			regex := wildcardToRegex(tt.pattern)
			matched := matchWildcardPattern(regex, tt.test)
			if matched != tt.wantMatch {
				t.Errorf("wildcard %q matched %q: got %v, want %v", tt.pattern, tt.test, matched, tt.wantMatch)
			}
		})
	}
}

func TestFilterByConditions(t *testing.T) {
	uid0 := 0
	uid1000 := 1000
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", LocalPort: "22", LocalAddress: "0.0.0.0",
			PeerPort: "*", PeerAddress: "0.0.0.0",
		},
		{
			Netid: "tcp", State: "LISTEN", LocalPort: "80", LocalAddress: "0.0.0.0",
			PeerPort: "*", PeerAddress: "0.0.0.0",
			UID: &uid0,
		},
		{
			Netid: "tcp", State: "ESTAB", LocalPort: "8080", LocalAddress: "192.168.1.1",
			PeerPort: "443", PeerAddress: "10.0.0.1",
			UID: &uid1000,
		},
		{
			Netid: "udp", State: "UNCONN", LocalPort: "53", LocalAddress: "127.0.0.1",
			PeerPort: "*", PeerAddress: "0.0.0.0",
			UID: &uid0,
		},
	}

	tests := []struct {
		name       string
		conditions []string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "single condition single value",
			conditions: []string{"netid=tcp"},
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "single condition multiple values",
			conditions: []string{"port=22,80"},
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "multiple conditions AND logic",
			conditions: []string{"netid=tcp", "port=22,80"},
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "three conditions",
			conditions: []string{"netid=tcp", "state=LISTEN", "port=22"},
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:       "empty conditions",
			conditions: []string{},
			wantCount:  4,
			wantErr:    false,
		},
		{
			name:       "no matches",
			conditions: []string{"netid=tcp", "port=5000"},
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "multiple values with AND",
			conditions: []string{"netid=tcp", "uid=0,1000"},
			wantCount:  2,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterByConditions(entries, tt.conditions)
			if (err != nil) != tt.wantErr {
				t.Errorf("FilterByConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(result) != tt.wantCount {
				t.Errorf("FilterByConditions() got %d results, want %d", len(result), tt.wantCount)
			}
		})
	}
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func matchWildcardPattern(regex, test string) bool {
	matched, _ := regexp.MatchString("(?i)"+regex, test)
	return matched
}
