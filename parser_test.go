package uss

import (
	"strings"
	"testing"
)

func TestParse_INET(t *testing.T) {
	input := `Netid  State    Recv-Q   Send-Q                           Local Address:Port      Peer Address:Port  Process
udp    UNCONN   0        0                                      0.0.0.0:36052          0.0.0.0:*      ino:408843379 sk:1001 cgroup:unreachable:64687d <->
udp    UNCONN   0        0                                127.0.0.53%lo:53             0.0.0.0:*      users:(("systemd-resolve",pid=1186020,fd=14)) uid:991 ino:408830904 sk:100a cgroup:/system.slice/systemd-resolved.service <->
tcp    LISTEN   0        4096                                   0.0.0.0:22             0.0.0.0:*      users:(("sshd",pid=1191350,fd=3),("systemd",pid=1,fd=92)) ino:148777152 sk:1028 cgroup:/system.slice/ssh.socket <->
tcp    LISTEN   0        32         [fe80::1266:6aff:fecf:96d]%incusbr0:53                [::]:*      users:(("dnsmasq",pid=1986269,fd=11)) ino:676938787 sk:600b cgroup:/system.slice/incus.service v6only:1 <->
tcp    LISTEN   0        4096                                    *:8090                   *:*      users:(("haproxy",pid=3558736,fd=18)) ino:606074948 sk:103b cgroup:/system.slice/haproxy.service v6only:0 <->
`

	entries, err := Parse(strings.NewReader(input), Options{Strict: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("Expected 5 entries, got %d", len(entries))
	}

	// Test first entry - basic UDP
	e := entries[0]
	if e.Netid != "udp" {
		t.Errorf("Entry 0: expected netid 'udp', got %q", e.Netid)
	}
	if e.State != "UNCONN" {
		t.Errorf("Entry 0: expected state 'UNCONN', got %q", e.State)
	}
	if e.LocalAddress != "0.0.0.0" {
		t.Errorf("Entry 0: expected local address '0.0.0.0', got %q", e.LocalAddress)
	}
	if e.LocalPort != "36052" {
		t.Errorf("Entry 0: expected local port '36052', got %q", e.LocalPort)
	}
	if e.Inode == nil || *e.Inode != 408843379 {
		t.Errorf("Entry 0: expected inode 408843379, got %v", e.Inode)
	}

	// Test second entry - UDP with interface and user
	e = entries[1]
	if e.LocalAddress != "127.0.0.53" {
		t.Errorf("Entry 1: expected local address '127.0.0.53', got %q", e.LocalAddress)
	}
	if e.Interface != "lo" {
		t.Errorf("Entry 1: expected interface 'lo', got %q", e.Interface)
	}
	if e.LocalPort != "53" {
		t.Errorf("Entry 1: expected local port '53', got %q", e.LocalPort)
	}
	if len(e.Users) != 1 {
		t.Fatalf("Entry 1: expected 1 user, got %d", len(e.Users))
	}
	if e.Users[0].Name != "systemd-resolve" {
		t.Errorf("Entry 1: expected user 'systemd-resolve', got %q", e.Users[0].Name)
	}
	if e.Users[0].PID != 1186020 {
		t.Errorf("Entry 1: expected PID 1186020, got %d", e.Users[0].PID)
	}
	if e.UID == nil || *e.UID != 991 {
		t.Errorf("Entry 1: expected UID 991, got %v", e.UID)
	}

	// Test third entry - TCP with multiple users
	e = entries[2]
	if e.Netid != "tcp" {
		t.Errorf("Entry 2: expected netid 'tcp', got %q", e.Netid)
	}
	if len(e.Users) != 2 {
		t.Fatalf("Entry 2: expected 2 users, got %d", len(e.Users))
	}

	// Test fourth entry - IPv6 with interface
	e = entries[3]
	if e.LocalAddress != "[fe80::1266:6aff:fecf:96d]" {
		t.Errorf("Entry 3: expected IPv6 address '[fe80::1266:6aff:fecf:96d]', got %q", e.LocalAddress)
	}
	if e.Interface != "incusbr0" {
		t.Errorf("Entry 3: expected interface 'incusbr0', got %q", e.Interface)
	}
	if e.V6Only == nil || *e.V6Only != 1 {
		t.Errorf("Entry 3: expected v6only 1, got %v", e.V6Only)
	}

	// Test fifth entry - wildcard address
	e = entries[4]
	if e.LocalAddress != "*" {
		t.Errorf("Entry 4: expected local address '*', got %q", e.LocalAddress)
	}
	if e.LocalPort != "8090" {
		t.Errorf("Entry 4: expected local port '8090', got %q", e.LocalPort)
	}
	if e.V6Only == nil || *e.V6Only != 0 {
		t.Errorf("Entry 4: expected v6only 0, got %v", e.V6Only)
	}
}

func TestParse_UNIX(t *testing.T) {
	input := `Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process
u_str LISTEN 0      4096   /run/tailscale/tailscaled.sock 632387454            * 0    users:(("tailscaled",pid=2922377,fd=11)) <-> ino:117522 dev:0/26 peers:
u_dgr UNCONN 0      0      /run/user/1000/systemd/notify 531040107             * 0    users:(("systemd",pid=715223,fd=18)) <-> ino:21 dev:1/12
u_str LISTEN 0      4096   @/var/lib/incus/containers/satis/command 123320627  * 0    users:(("incusd",pid=2616744,fd=8)) <-> peers:
u_seq LISTEN 0      4096   /run/udev/control 16499            * 0
u_str LISTEN 0      128    @5cbfc 47960                                         * 0    users:(("incusd",pid=380945,fd=10),("incusd",pid=147286,fd=10)) <-> peers:
`

	entries, err := Parse(strings.NewReader(input), Options{Strict: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(entries) != 5 {
		t.Fatalf("Expected 5 entries, got %d", len(entries))
	}

	// Test first entry - basic UNIX stream socket
	e := entries[0]
	if e.Netid != "u_str" {
		t.Errorf("Entry 0: expected netid 'u_str', got %q", e.Netid)
	}
	if e.UnixPath != "/run/tailscale/tailscaled.sock" {
		t.Errorf("Entry 0: expected path '/run/tailscale/tailscaled.sock', got %q", e.UnixPath)
	}
	if e.UnixID != "632387454" {
		t.Errorf("Entry 0: expected ID '632387454', got %q", e.UnixID)
	}
	if e.UnixPeer != "*" {
		t.Errorf("Entry 0: expected peer '*', got %q", e.UnixPeer)
	}
	if e.UnixPeerID != "0" {
		t.Errorf("Entry 0: expected peer ID '0', got %q", e.UnixPeerID)
	}
	if len(e.Users) != 1 {
		t.Fatalf("Entry 0: expected 1 user, got %d", len(e.Users))
	}

	// Test second entry - UNIX datagram
	e = entries[1]
	if e.Netid != "u_dgr" {
		t.Errorf("Entry 1: expected netid 'u_dgr', got %q", e.Netid)
	}

	// Test third entry - abstract socket
	e = entries[2]
	if e.UnixPath != "@/var/lib/incus/containers/satis/command" {
		t.Errorf("Entry 2: expected abstract socket path, got %q", e.UnixPath)
	}

	// Test fourth entry - short format (no users)
	e = entries[3]
	if e.Netid != "u_seq" {
		t.Errorf("Entry 3: expected netid 'u_seq', got %q", e.Netid)
	}
	if e.UnixPath != "/run/udev/control" {
		t.Errorf("Entry 3: expected path '/run/udev/control', got %q", e.UnixPath)
	}

	// Test fifth entry - short abstract socket with multiple users
	e = entries[4]
	if e.UnixPath != "@5cbfc" {
		t.Errorf("Entry 4: expected path '@5cbfc', got %q", e.UnixPath)
	}
	if len(e.Users) != 2 {
		t.Fatalf("Entry 4: expected 2 users, got %d", len(e.Users))
	}
}

func TestParse_SkipErrorLines(t *testing.T) {
	input := `Failed to open cgroup2 by ID
Failed to open cgroup2 by ID
Netid  State    Recv-Q   Send-Q   Local Address:Port   Peer Address:Port
tcp    LISTEN   0        4096     0.0.0.0:22          0.0.0.0:*
`

	entries, err := Parse(strings.NewReader(input), Options{Strict: true})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry (errors should be skipped), got %d", len(entries))
	}

	if entries[0].Netid != "tcp" {
		t.Errorf("Expected netid 'tcp', got %q", entries[0].Netid)
	}
}

func TestSplitAddressPort(t *testing.T) {
	tests := []struct {
		input     string
		wantAddr  string
		wantPort  string
		wantIface string
	}{
		{"0.0.0.0:111", "0.0.0.0", "111", ""},
		{"127.0.0.53%lo:53", "127.0.0.53", "53", "lo"},
		{"[::]:22", "[::]", "22", ""},
		{"[fe80::1266:6aff:fecf:96d]%incusbr0:53", "[fe80::1266:6aff:fecf:96d]", "53", "incusbr0"},
		{"*:989", "*", "989", ""},
		{"*", "*", "*", ""},
		{":*", "", "*", ""},
		{"_:8090_", "_", "8090", ""},
		{"0.0.0.0%incusbr0:67", "0.0.0.0", "67", "incusbr0"},
		{"[fd42:2eff:9d69:efec::1]:53", "[fd42:2eff:9d69:efec::1]", "53", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			addr, port, iface := splitAddressPort(tt.input)
			if addr != tt.wantAddr {
				t.Errorf("address: got %q, want %q", addr, tt.wantAddr)
			}
			if port != tt.wantPort {
				t.Errorf("port: got %q, want %q", port, tt.wantPort)
			}
			if iface != tt.wantIface {
				t.Errorf("interface: got %q, want %q", iface, tt.wantIface)
			}
		})
	}
}
