package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	formatFlag string
	prettyFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "aical",
	Short: "CLI for Apple Calendar.app automation via Apple Events",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "", "output format: json")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "pretty-print JSON output")
}

// isJSON returns true when output should be JSON (--format json or stdout is not a TTY).
func isJSON(cmd *cobra.Command) bool {
	if formatFlag == "json" {
		return true
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return true
	}
	return false
}
