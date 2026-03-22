package cmd

import (
	"fmt"
	"strconv"

	"github.com/blakemcanally/apple-mail-cli/internal/mail"
	"github.com/blakemcanally/apple-mail-cli/internal/output"
	"github.com/spf13/cobra"
)

var mailboxesAccountFlag string

var mailboxesCmd = &cobra.Command{
	Use:   "mailboxes",
	Short: "List mailboxes/folders",
	RunE: func(cmd *cobra.Command, args []string) error {
		mailboxes, err := mail.ListMailboxes(mailboxesAccountFlag)
		if err != nil {
			output.PrintError("mailboxes", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("mailboxes", mailboxes, prettyFlag)
			return nil
		}

		t := output.NewTable("NAME", "ACCOUNT", "UNREAD", "TOTAL")
		for _, m := range mailboxes {
			t.AddRow(m.Name, m.Account, strconv.Itoa(m.UnreadCount), strconv.Itoa(m.TotalCount))
		}
		fmt.Printf("%d mailbox(es)\n\n", len(mailboxes))
		t.Print()
		return nil
	},
}

func init() {
	mailboxesCmd.Flags().StringVar(&mailboxesAccountFlag, "account", "", "filter by account name")
	rootCmd.AddCommand(mailboxesCmd)
}
