package uss

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MatchType defines the type of matching strategy
type MatchType int

const (
	MatchExact MatchType = iota
	MatchWildcard
	MatchSubstring
	MatchRange
)

// ValueMatcher represents a single value matching strategy
type ValueMatcher struct {
	Type  MatchType
	Value string
	Min   *int
	Max   *int
}

// FieldCondition represents conditions for a single field
type FieldCondition struct {
	Field    string
	Matchers []ValueMatcher
}

// FilterQuery represents a parsed query
type FilterQuery struct {
	Conditions []FieldCondition
}

// Filter filters entries based on a query string
// Query format: field=value1,value2 field2=value3 (space or semicolon separated)
// Multiple values per field are OR'd, multiple fields are AND'd
func Filter(entries []Entry, query string) ([]Entry, error) {
	if strings.TrimSpace(query) == "" {
		return entries, nil
	}

	filterQuery, err := parseQuery(query)
	if err != nil {
		return nil, err
	}

	var result []Entry
	for _, entry := range entries {
		if matchEntry(entry, filterQuery) {
			result = append(result, entry)
		}
	}
	return result, nil
}

// FilterByConditions filters entries based on a slice of condition strings
// Each condition string is in format "field=value1,value2"
// Multiple conditions are AND'd together
// Example: FilterByConditions(entries, []string{"port=80,22", "netid=tcp", "state=LISTEN"})
func FilterByConditions(entries []Entry, conditions []string) ([]Entry, error) {
	if len(conditions) == 0 {
		return entries, nil
	}

	// Join conditions with space and use existing Filter function
	query := strings.Join(conditions, " ")
	return Filter(entries, query)
}

// parseQuery parses a query string into FilterQuery
func parseQuery(query string) (FilterQuery, error) {
	// Split by space and semicolon
	conditions := splitQuery(query)
	if len(conditions) == 0 {
		return FilterQuery{}, nil
	}

	var filterQuery FilterQuery
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}

		fieldCond, err := parseFieldCondition(condition)
		if err != nil {
			return FilterQuery{}, err
		}
		filterQuery.Conditions = append(filterQuery.Conditions, fieldCond)
	}

	return filterQuery, nil
}

// splitQuery splits query by space and/or semicolon separators
func splitQuery(query string) []string {
	// Replace semicolon with space, then split by spaces
	normalized := strings.ReplaceAll(query, ";", " ")
	return strings.Fields(normalized)
}

// parseFieldCondition parses "field=value1,value2" into FieldCondition
func parseFieldCondition(condition string) (FieldCondition, error) {
	parts := strings.SplitN(condition, "=", 2)
	if len(parts) != 2 {
		return FieldCondition{}, fmt.Errorf("invalid condition format: %s (expected field=value)", condition)
	}

	field := strings.TrimSpace(parts[0])
	valuesStr := parts[1]

	if field == "" {
		return FieldCondition{}, fmt.Errorf("field name cannot be empty in: %s", condition)
	}

	// Normalize field name to lowercase
	field = strings.ToLower(field)

	// Split values by comma
	values := strings.Split(valuesStr, ",")
	if len(values) == 0 {
		return FieldCondition{}, fmt.Errorf("no values specified for field: %s", field)
	}

	var matchers []ValueMatcher
	for _, val := range values {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		// Determine field type for parsing
		fieldType := getFieldType(field)
		matcher, err := parseValue(val, fieldType)
		if err != nil {
			return FieldCondition{}, fmt.Errorf("error parsing value %q for field %s: %w", val, field, err)
		}
		matchers = append(matchers, matcher)
	}

	if len(matchers) == 0 {
		return FieldCondition{}, fmt.Errorf("no valid values specified for field: %s", field)
	}

	return FieldCondition{
		Field:    field,
		Matchers: matchers,
	}, nil
}

// parseValue parses a single value into ValueMatcher (detects type)
func parseValue(value, fieldType string) (ValueMatcher, error) {
	// Check for range (numeric fields only)
	if fieldType == "numeric" && strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) == 2 {
			min, errMin := strconv.Atoi(strings.TrimSpace(parts[0]))
			max, errMax := strconv.Atoi(strings.TrimSpace(parts[1]))
			if errMin == nil && errMax == nil && min <= max {
				return ValueMatcher{
					Type:  MatchRange,
					Value: value,
					Min:   &min,
					Max:   &max,
				}, nil
			}
		}
	}

	// Check for wildcard
	if strings.Contains(value, "*") || strings.Contains(value, "?") {
		return ValueMatcher{
			Type:  MatchWildcard,
			Value: value,
		}, nil
	}

	// Default to exact match, but substring is fallback during matching
	return ValueMatcher{
		Type:  MatchExact,
		Value: value,
	}, nil
}

// matchEntry checks if entry matches all conditions (AND logic)
func matchEntry(entry Entry, query FilterQuery) bool {
	for _, condition := range query.Conditions {
		if !matchFieldCondition(entry, condition) {
			return false
		}
	}
	return true
}

// matchFieldCondition checks if entry matches field condition (OR logic for multiple matchers)
func matchFieldCondition(entry Entry, cond FieldCondition) bool {
	// Handle special convenience fields that match multiple entry fields
	switch cond.Field {
	case "port":
		// Match either local or peer port
		localPort := parsePort(entry.LocalPort)
		peerPort := parsePort(entry.PeerPort)
		for _, matcher := range cond.Matchers {
			if (localPort != nil && matchValue(localPort, matcher, "numeric")) ||
				(peerPort != nil && matchValue(peerPort, matcher, "numeric")) {
				return true
			}
		}
		return false

	case "address":
		// Match either local or peer address
		localAddr := entry.LocalAddress
		peerAddr := entry.PeerAddress
		for _, matcher := range cond.Matchers {
			if (localAddr != "" && matchValue(localAddr, matcher, "string")) ||
				(peerAddr != "" && matchValue(peerAddr, matcher, "string")) {
				return true
			}
		}
		return false
	}

	// Regular field matching
	actualValue := getFieldValue(entry, cond.Field)
	if actualValue == nil {
		return false
	}

	fieldType := getFieldType(cond.Field)

	// If any matcher matches, the condition is satisfied (OR logic)
	for _, matcher := range cond.Matchers {
		if matchValue(actualValue, matcher, fieldType) {
			return true
		}
	}
	return false
}

// matchValue performs actual matching based on ValueMatcher type
// fieldType is "numeric" or "string" to determine if substring matching is allowed
func matchValue(actualValue interface{}, matcher ValueMatcher, fieldType string) bool {
	switch matcher.Type {
	case MatchRange:
		if numVal, ok := actualValue.(int); ok {
			if matcher.Min != nil && matcher.Max != nil {
				return numVal >= *matcher.Min && numVal <= *matcher.Max
			}
		}
		return false

	case MatchWildcard:
		strVal := toString(actualValue)
		// Convert wildcard pattern to regex
		pattern := wildcardToRegex(matcher.Value)
		matched, _ := regexp.MatchString("(?i)"+pattern, strVal) // case-insensitive
		return matched

	case MatchExact:
		strVal := toString(actualValue)
		matchStr := matcher.Value

		// Normalize to lowercase for case-insensitive comparison
		strVal = strings.ToLower(strVal)
		matchStr = strings.ToLower(matchStr)

		// Try exact match first
		if strVal == matchStr {
			return true
		}

		// For string fields only, fallback to substring match
		if fieldType == "string" {
			return strings.Contains(strVal, matchStr)
		}

		// For numeric fields, only allow exact match
		return false

	default:
		return false
	}
}

// getFieldValue retrieves the value of a field from an entry
func getFieldValue(entry Entry, field string) interface{} {
	switch field {
	// Common fields
	case "netid":
		return entry.Netid
	case "state":
		return entry.State
	case "recvq":
		return entry.RecvQ
	case "sendq":
		return entry.SendQ

	// INET fields
	case "localaddress":
		return entry.LocalAddress
	case "localport":
		return parsePort(entry.LocalPort)
	case "peeraddress":
		return entry.PeerAddress
	case "peerport":
		return parsePort(entry.PeerPort)
	case "interface":
		return entry.Interface

	// Convenience fields
	case "port":
		// Match either local or peer port
		return nil // Handled specially
	case "address":
		// Match either local or peer address
		return nil // Handled specially

	// UNIX fields
	case "unixtype":
		return entry.UnixType
	case "unixpath":
		return entry.UnixPath
	case "unixid":
		return entry.UnixID
	case "unixpeer":
		return entry.UnixPeer
	case "unixpeerid":
		return entry.UnixPeerID

	// Metadata fields
	case "uid":
		if entry.UID != nil {
			return *entry.UID
		}
		return nil
	case "inode":
		if entry.Inode != nil {
			return int(*entry.Inode)
		}
		return nil
	case "cgroup":
		return entry.CGroup
	case "v6only":
		if entry.V6Only != nil {
			return *entry.V6Only
		}
		return nil
	case "fwmark":
		return entry.FWMark
	case "sk":
		return entry.Sk
	case "dev":
		return entry.Dev

	default:
		return nil
	}
}

// getFieldType returns the type of a field for parsing purposes
func getFieldType(field string) string {
	numericFields := []string{
		"recvq", "sendq", "localport", "peerport", "port",
		"uid", "inode", "v6only",
	}

	for _, nf := range numericFields {
		if field == nf {
			return "numeric"
		}
	}
	return "string"
}

// parsePort converts a port string to int, returning nil if invalid
func parsePort(portStr string) interface{} {
	if portStr == "" || portStr == "*" {
		return nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil
	}
	return port
}

// toString converts a value to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case uint64:
		return strconv.FormatUint(val, 10)
	case *string:
		if val == nil {
			return ""
		}
		return *val
	case *int:
		if val == nil {
			return ""
		}
		return strconv.Itoa(*val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// wildcardToRegex converts wildcard pattern to regex
func wildcardToRegex(pattern string) string {
	// Escape regex special characters except * and ?
	escaped := regexp.QuoteMeta(pattern)
	// Replace escaped * with .* (any characters)
	escaped = strings.ReplaceAll(escaped, `\*`, ".*")
	// Replace escaped ? with . (any single character)
	escaped = strings.ReplaceAll(escaped, `\?`, ".")
	return escaped
}
