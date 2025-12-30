package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/mobydeck/uss"
	"github.com/spf13/cobra"
)

var (
	flagCSV    bool
	flagJSON   bool
	flagYAML   bool
	flagPretty bool
	flagQuery  string
	flagRaw    bool
	flagSS     string
)

var rootCmd = &cobra.Command{
	Use:   "uss [file]",
	Short: "Parse and format Linux socket statistics (ss output)",
	Long: `uss parses Linux socket statistics from ss command output.
It supports both INET (tcp/udp) and UNIX sockets, and can output
in multiple formats: human-readable (default), CSV, JSON, or YAML.

By default, reads from stdin. Optionally accepts a filename argument.`,
	Example: `  # Read from stdin, human-readable output
  ss -tunap | uss

  # Read from stdin, JSON output
  ss -tunap | uss --json

  # Execute ss directly with flags
  uss --ss "-tunap"

  # Execute ss and output as JSON
  uss --ss "-peanutlx" --json

  # Read from file, YAML output
  ss -peanutlx > sockets.txt
  uss --yaml sockets.txt

  # Filter results
  ss -tunap | uss -q "port=22,80"`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().BoolVarP(&flagCSV, "csv", "c", false, "Output as CSV")
	rootCmd.Flags().BoolVarP(&flagJSON, "json", "j", false, "Output as JSON")
	rootCmd.Flags().BoolVarP(&flagYAML, "yaml", "y", false, "Output as YAML")
	rootCmd.Flags().BoolVarP(&flagPretty, "pretty", "p", false, "Pretty-print JSON (only valid with --json)")
	rootCmd.Flags().StringVarP(&flagQuery, "query", "q", "", "Filter results (e.g., -q \"port=22,80 uid=0\")")
	rootCmd.Flags().BoolVarP(&flagRaw, "raw", "r", false, "Include raw unparsed fields (processRaw, local, peer)")
	rootCmd.Flags().StringVar(&flagSS, "ss", "", "Execute ss with these flags (e.g., --ss \"-tunap\")")
}

func run(cmd *cobra.Command, args []string) error {
	// Validate flags
	formatCount := 0
	if flagCSV {
		formatCount++
	}
	if flagJSON {
		formatCount++
	}
	if flagYAML {
		formatCount++
	}

	if formatCount > 1 {
		return fmt.Errorf("only one output format flag (--csv, --json, --yaml) may be specified")
	}

	if flagPretty && !flagJSON {
		return fmt.Errorf("--pretty flag requires --json")
	}

	if flagSS != "" && len(args) > 0 {
		return fmt.Errorf("--ss flag cannot be used with a file argument")
	}

	// Determine input source
	var input io.ReadCloser
	var err error

	if flagSS != "" {
		// Execute ss with provided flags
		// Add leading dash if not present
		ssFlags := flagSS
		if !strings.HasPrefix(ssFlags, "-") {
			ssFlags = "-" + ssFlags
		}

		ssCmd := exec.Command("ss", ssFlags)
		input, err = ssCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("creating ss pipe: %w", err)
		}
		defer input.Close()

		if err := ssCmd.Start(); err != nil {
			return fmt.Errorf("executing ss: %w", err)
		}
		defer ssCmd.Wait()
	} else if len(args) == 0 {
		// Read from stdin
		input = os.Stdin
	} else {
		// Read from file
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		input = file
	}
	defer input.Close()

	// Parse input
	entries, err := uss.Parse(input, uss.Options{Strict: false})
	if err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}

	// Apply filtering if query provided
	if flagQuery != "" {
		entries, err = uss.Filter(entries, flagQuery)
		if err != nil {
			return fmt.Errorf("filtering results: %w", err)
		}
	}

	// Render output
	if flagCSV {
		return uss.RenderCSV(os.Stdout, entries, flagRaw)
	} else if flagJSON {
		return uss.RenderJSON(os.Stdout, entries, flagPretty, flagRaw)
	} else if flagYAML {
		return uss.RenderYAML(os.Stdout, entries, flagRaw)
	}

	// Default: human-readable
	return uss.RenderHuman(os.Stdout, entries, flagRaw)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
