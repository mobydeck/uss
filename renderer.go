package uss

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"go.yaml.in/yaml/v4"
)

// RenderHuman renders entries in a human-readable table format
func RenderHuman(w io.Writer, entries []Entry, showRaw bool) error {
	if len(entries) == 0 {
		return nil
	}

	if !showRaw {
		entries = cleanEntries(entries)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()

	// Determine if we have INET or UNIX sockets (or mixed)
	hasINET := false
	hasUNIX := false
	for _, e := range entries {
		if strings.HasPrefix(e.Netid, "u_") {
			hasUNIX = true
		} else {
			hasINET = true
		}
	}

	// Render INET entries
	if hasINET {
		fmt.Fprintf(tw, "Netid\tState\tRecv-Q\tSend-Q\tLocal\tPeer\tProcess\n")
		for _, e := range entries {
			if strings.HasPrefix(e.Netid, "u_") {
				continue
			}
			local := e.Local
			if local == "" && e.LocalAddress != "" {
				local = e.LocalAddress + ":" + e.LocalPort
			}
			peer := e.Peer
			if peer == "" && e.PeerAddress != "" {
				peer = e.PeerAddress + ":" + e.PeerPort
			}
			process := truncateProcess(e.ProcessRaw, 50)
			fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\t%s\t%s\n",
				e.Netid, e.State, e.RecvQ, e.SendQ, local, peer, process)
		}
	}

	// Add separator if both types exist
	if hasINET && hasUNIX {
		fmt.Fprintln(tw)
	}

	// Render UNIX entries
	if hasUNIX {
		fmt.Fprintf(tw, "Netid\tState\tRecv-Q\tSend-Q\tPath\tID\tPeer\tProcess\n")
		for _, e := range entries {
			if !strings.HasPrefix(e.Netid, "u_") {
				continue
			}
			process := truncateProcess(e.ProcessRaw, 50)
			fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\t%s\t%s\t%s\n",
				e.Netid, e.State, e.RecvQ, e.SendQ, e.UnixPath, e.UnixID, e.UnixPeer, process)
		}
	}

	return nil
}

// truncateProcess truncates a process string to maxLen
func truncateProcess(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// RenderCSV renders entries as CSV
func RenderCSV(w io.Writer, entries []Entry, showRaw bool) error {
	if !showRaw {
		entries = cleanEntries(entries)
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	header := []string{
		"Netid", "State", "RecvQ", "SendQ",
		"LocalAddress", "LocalPort", "PeerAddress", "PeerPort",
		"Interface",
		"UnixPath", "UnixID", "UnixPeer", "UnixPeerID",
		"Users", "UID", "Inode", "CGroup", "V6Only", "FWMark", "Sk", "Dev", "Peers",
	}
	if showRaw {
		header = append(header, "ProcessRaw")
	}
	if err := cw.Write(header); err != nil {
		return err
	}

	// Write entries
	for _, e := range entries {
		usersJSON, _ := json.Marshal(e.Users)
		row := []string{
			e.Netid,
			e.State,
			fmt.Sprintf("%d", e.RecvQ),
			fmt.Sprintf("%d", e.SendQ),
			e.LocalAddress,
			e.LocalPort,
			e.PeerAddress,
			e.PeerPort,
			e.Interface,
			e.UnixPath,
			e.UnixID,
			e.UnixPeer,
			e.UnixPeerID,
			string(usersJSON),
			ptrIntToString(e.UID),
			ptrUint64ToString(e.Inode),
			ptrStringToString(e.CGroup),
			ptrIntToString(e.V6Only),
			ptrStringToString(e.FWMark),
			ptrStringToString(e.Sk),
			ptrStringToString(e.Dev),
			ptrStringToString(e.Peers),
		}
		if showRaw {
			row = append(row, e.ProcessRaw)
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// RenderJSON renders entries as JSON
func RenderJSON(w io.Writer, entries []Entry, pretty bool, showRaw bool) error {
	if !showRaw {
		entries = cleanEntries(entries)
	}

	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(entries)
}

// RenderYAML renders entries as YAML
func RenderYAML(w io.Writer, entries []Entry, showRaw bool) error {
	if !showRaw {
		entries = cleanEntries(entries)
	}

	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(entries)
}

// cleanEntries removes raw/unparsed fields from entries for cleaner output
func cleanEntries(entries []Entry) []Entry {
	cleaned := make([]Entry, len(entries))
	for i, e := range entries {
		cleaned[i] = e
		cleaned[i].ProcessRaw = ""
		cleaned[i].Local = ""
		cleaned[i].Peer = ""
	}
	return cleaned
}

// Helper functions for CSV rendering
func ptrIntToString(p *int) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d", *p)
}

func ptrUint64ToString(p *uint64) string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("%d", *p)
}

func ptrStringToString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
