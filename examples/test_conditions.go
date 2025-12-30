package main

import (
	"fmt"
	"strings"

	"github.com/mobydeck/uss"
)

func main() {
	input := `Netid  State    Recv-Q   Send-Q   Local Address:Port   Peer Address:Port
tcp    LISTEN   0        128      0.0.0.0:22              0.0.0.0:*
tcp    LISTEN   0        128      0.0.0.0:80              0.0.0.0:*
tcp    LISTEN   0        128      0.0.0.0:443             0.0.0.0:*
udp    UNCONN   0        0        127.0.0.53%lo:53        0.0.0.0:*
`

	entries, _ := uss.Parse(strings.NewReader(input), uss.Options{Strict: false})

	fmt.Println("=== Test 1: Single condition ===")
	result, _ := uss.FilterByConditions(entries, []string{"netid=tcp"})
	fmt.Printf("netid=tcp: %d results\n", len(result))

	fmt.Println("\n=== Test 2: Multiple conditions (AND) ===")
	result, _ = uss.FilterByConditions(entries, []string{"netid=tcp", "port=22,80"})
	fmt.Printf("netid=tcp AND port=22,80: %d results\n", len(result))
	for _, e := range result {
		fmt.Printf("  - %s:%s\n", e.LocalAddress, e.LocalPort)
	}

	fmt.Println("\n=== Test 3: Empty conditions ===")
	result, _ = uss.FilterByConditions(entries, []string{})
	fmt.Printf("Empty conditions: %d results (all entries)\n", len(result))

	fmt.Println("\n=== Test 4: Complex conditions ===")
	result, _ = uss.FilterByConditions(entries, []string{"netid=tcp", "state=LISTEN", "port=80,443"})
	fmt.Printf("netid=tcp AND state=LISTEN AND port=80,443: %d results\n", len(result))
	for _, e := range result {
		fmt.Printf("  - %s:%s (%s)\n", e.LocalAddress, e.LocalPort, e.State)
	}
}
