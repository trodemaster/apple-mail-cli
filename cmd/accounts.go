package cmd

import (
	"fmt"
	"strings"

	"github.com/blakemcanally/apple-mail-cli/internal/mail"
	"github.com/blakemcanally/apple-mail-cli/internal/output"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List configured mail accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		accounts, err := mail.ListAccounts()
		if err != nil {
			output.PrintError("accounts", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("accounts", accounts, prettyFlag)
			return nil
		}

		t := output.NewTable("NAME", "EMAILS", "TYPE")
		for _, a := range accounts {
			t.AddRow(a.Name, strings.Join(a.EmailAddresses, ", "), a.Type)
		}
		fmt.Printf("%d account(s)\n\n", len(accounts))
		t.Print()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}
