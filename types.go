package uss

// User represents a process user extracted from ss output
type User struct {
	Name string `json:"name" yaml:"name"`
	PID  int    `json:"pid" yaml:"pid"`
	FD   int    `json:"fd" yaml:"fd"`
}

// Entry represents a single socket entry from ss output
type Entry struct {
	// Common fields
	Netid      string `json:"netid" yaml:"netid"`
	State      string `json:"state" yaml:"state"`
	RecvQ      int    `json:"recvQ" yaml:"recvQ"`
	SendQ      int    `json:"sendQ" yaml:"sendQ"`
	ProcessRaw string `json:"processRaw,omitempty" yaml:"processRaw,omitempty"`

	// Extracted process metadata
	Users  []User  `json:"users,omitempty" yaml:"users,omitempty"`
	UID    *int    `json:"uid,omitempty" yaml:"uid,omitempty"`
	Inode  *uint64 `json:"inode,omitempty" yaml:"inode,omitempty"`
	CGroup *string `json:"cgroup,omitempty" yaml:"cgroup,omitempty"`
	V6Only *int    `json:"v6only,omitempty" yaml:"v6only,omitempty"` // 0 or 1
	FWMark *string `json:"fwmark,omitempty" yaml:"fwmark,omitempty"`
	Sk     *string `json:"sk,omitempty" yaml:"sk,omitempty"`
	Dev    *string `json:"dev,omitempty" yaml:"dev,omitempty"`
	Peers  *string `json:"peers,omitempty" yaml:"peers,omitempty"`

	// INET-specific fields
	LocalAddress string `json:"localAddress,omitempty" yaml:"localAddress,omitempty"`
	LocalPort    string `json:"localPort,omitempty" yaml:"localPort,omitempty"`
	PeerAddress  string `json:"peerAddress,omitempty" yaml:"peerAddress,omitempty"`
	PeerPort     string `json:"peerPort,omitempty" yaml:"peerPort,omitempty"`
	Local        string `json:"local,omitempty" yaml:"local,omitempty"` // raw token
	Peer         string `json:"peer,omitempty" yaml:"peer,omitempty"`   // raw token
	Interface    string `json:"interface,omitempty" yaml:"interface,omitempty"`

	// UNIX-specific fields
	UnixType   string `json:"unixType,omitempty" yaml:"unixType,omitempty"`
	UnixPath   string `json:"unixPath,omitempty" yaml:"unixPath,omitempty"`
	UnixID     string `json:"unixId,omitempty" yaml:"unixId,omitempty"`
	UnixPeer   string `json:"unixPeer,omitempty" yaml:"unixPeer,omitempty"`
	UnixPeerID string `json:"unixPeerId,omitempty" yaml:"unixPeerId,omitempty"`
}

// Options controls parsing behavior
type Options struct {
	Strict bool // if true, fail on malformed lines; if false, skip and continue
}
