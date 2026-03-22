package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type flagSpec struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
	Required  bool   `json:"required,omitempty"`
}

type commandSpec struct {
	Use         string        `json:"use"`
	Name        string        `json:"name"`
	Short       string        `json:"short"`
	Long        string        `json:"long,omitempty"`
	Flags       []flagSpec    `json:"flags,omitempty"`
	Subcommands []commandSpec `json:"subcommands,omitempty"`
}

func buildSpec(c *cobra.Command) commandSpec {
	spec := commandSpec{
		Use:   c.Use,
		Name:  c.Name(),
		Short: c.Short,
		Long:  c.Long,
	}

	c.Flags().VisitAll(func(f *pflag.Flag) {
		fs := flagSpec{
			Name:    f.Name,
			Type:    f.Value.Type(),
			Default: f.DefValue,
			Usage:   f.Usage,
		}
		if f.Shorthand != "" {
			fs.Shorthand = f.Shorthand
		}
		if anno, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(anno) > 0 {
			fs.Required = true
		}
		spec.Flags = append(spec.Flags, fs)
	})

	// include inherited persistent flags (e.g. --format, --pretty)
	c.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		fs := flagSpec{
			Name:    f.Name,
			Type:    f.Value.Type(),
			Default: f.DefValue,
			Usage:   f.Usage,
		}
		if f.Shorthand != "" {
			fs.Shorthand = f.Shorthand
		}
		spec.Flags = append(spec.Flags, fs)
	})

	for _, sub := range c.Commands() {
		if sub.Name() == "help" || sub.Name() == "completion" {
			continue
		}
		spec.Subcommands = append(spec.Subcommands, buildSpec(sub))
	}

	return spec
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print a JSON schema of all commands and flags (agent-friendly)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root := buildSpec(cmd.Root())
		// remove schema itself from output — not useful to agents
		filtered := make([]commandSpec, 0, len(root.Subcommands))
		for _, s := range root.Subcommands {
			if s.Name != "schema" {
				filtered = append(filtered, s)
			}
		}
		root.Subcommands = filtered

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(root); err != nil {
			return fmt.Errorf("encode schema: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
