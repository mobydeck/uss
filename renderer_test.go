package uss

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderHuman(t *testing.T) {
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", RecvQ: 0, SendQ: 128,
			Local: "0.0.0.0:22", Peer: "0.0.0.0:*",
			LocalAddress: "0.0.0.0", LocalPort: "22",
			PeerAddress: "0.0.0.0", PeerPort: "*",
			ProcessRaw: "users:((\"sshd\",pid=1234,fd=3))",
		},
		{
			Netid: "u_str", State: "LISTEN", RecvQ: 0, SendQ: 4096,
			UnixPath: "/run/test.sock", UnixID: "12345", UnixPeer: "*",
			ProcessRaw: "users:((\"test\",pid=5678,fd=7))",
		},
	}

	var buf bytes.Buffer
	err := RenderHuman(&buf, entries, false)
	if err != nil {
		t.Fatalf("RenderHuman failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tcp") {
		t.Errorf("Output should contain 'tcp'")
	}
	if !strings.Contains(output, "u_str") {
		t.Errorf("Output should contain 'u_str'")
	}
	if !strings.Contains(output, "0.0.0.0:22") {
		t.Errorf("Output should contain '0.0.0.0:22'")
	}
	if !strings.Contains(output, "/run/test.sock") {
		t.Errorf("Output should contain '/run/test.sock'")
	}
}

func TestRenderCSV(t *testing.T) {
	uid := 1000
	inode := uint64(12345)
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", RecvQ: 0, SendQ: 128,
			LocalAddress: "0.0.0.0", LocalPort: "22",
			PeerAddress: "0.0.0.0", PeerPort: "*",
			UID:   &uid,
			Inode: &inode,
			Users: []User{{Name: "sshd", PID: 1234, FD: 3}},
		},
	}

	var buf bytes.Buffer
	err := RenderCSV(&buf, entries, false)
	if err != nil {
		t.Fatalf("RenderCSV failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines (header + 1 entry), got %d", len(lines))
	}

	// Check header
	if !strings.Contains(lines[0], "Netid") {
		t.Errorf("Header should contain 'Netid'")
	}

	// Check data
	if !strings.Contains(lines[1], "tcp") {
		t.Errorf("Data should contain 'tcp'")
	}
	if !strings.Contains(lines[1], "1000") {
		t.Errorf("Data should contain UID '1000'")
	}
}

func TestRenderJSON(t *testing.T) {
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", RecvQ: 0, SendQ: 128,
			LocalAddress: "0.0.0.0", LocalPort: "22",
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, entries, false, false)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\"netid\":\"tcp\"") {
		t.Errorf("JSON should contain '\"netid\":\"tcp\"'")
	}
}

func TestRenderJSON_Pretty(t *testing.T) {
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", RecvQ: 0, SendQ: 128,
			LocalAddress: "0.0.0.0", LocalPort: "22",
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, entries, true, false)
	if err != nil {
		t.Fatalf("RenderJSON (pretty) failed: %v", err)
	}

	output := buf.String()
	// Pretty JSON should have indentation
	if !strings.Contains(output, "\n  ") {
		t.Errorf("Pretty JSON should have indentation")
	}
}

func TestRenderYAML(t *testing.T) {
	entries := []Entry{
		{
			Netid: "tcp", State: "LISTEN", RecvQ: 0, SendQ: 128,
			LocalAddress: "0.0.0.0", LocalPort: "22",
		},
	}

	var buf bytes.Buffer
	err := RenderYAML(&buf, entries, false)
	if err != nil {
		t.Fatalf("RenderYAML failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "netid: tcp") {
		t.Errorf("YAML should contain 'netid: tcp'")
	}
}
