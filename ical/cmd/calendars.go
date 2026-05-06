package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/trodemaster/apple-mail-cli/internal/ical"
	"github.com/trodemaster/apple-mail-cli/internal/output"
)

var calendarsCmd = &cobra.Command{
	Use:   "calendars",
	Short: "List all calendars",
	RunE: func(cmd *cobra.Command, args []string) error {
		cals, err := ical.ListCalendars()
		if err != nil {
			output.PrintError("calendars", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("calendars", cals, prettyFlag)
			return nil
		}

		t := output.NewTable("NAME", "ID", "WRITABLE", "DESCRIPTION")
		for _, c := range cals {
			writable := "no"
			if c.Writable {
				writable = "yes"
			}
			t.AddRow(c.Name, c.ID, writable, c.Description)
		}
		fmt.Printf("%d calendar(s)\n\n", len(cals))
		t.Print()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(calendarsCmd)
}
