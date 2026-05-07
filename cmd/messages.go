package cmd

import (
	"fmt"
	"strings"

	"github.com/trodemaster/apple-mail-cli/internal/mail"
	"github.com/trodemaster/apple-mail-cli/internal/output"
	"github.com/spf13/cobra"
)

var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Message subcommands: list, read, search",
}

// --- list ---

var (
	msgListAccount string
	msgListMailbox string
	msgListLimit   int
	msgListUnread  bool
	msgListSummary bool
)

var messagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages in a mailbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		if msgListAccount == "" || msgListMailbox == "" {
			return fmt.Errorf("--account and --mailbox are required")
		}
		msgs, err := mail.ListMessages(mail.ListMessagesParams{
			AccountName: msgListAccount,
			MailboxName: msgListMailbox,
			Limit:       msgListLimit,
			UnreadOnly:  msgListUnread,
			IncludeBody: msgListSummary,
		})
		if err != nil {
			output.PrintError("messages list", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("messages list", msgs, prettyFlag)
			return nil
		}

		fmt.Printf("%d message(s)\n\n", len(msgs))
		if msgListSummary {
			printSummaryTable(msgs)
		} else {
			t := output.NewTable("SENDER", "SUBJECT", "DATE", "READ", "FLAGGED", "ATT")
			for _, m := range msgs {
				read, flagged, att := "no", "no", "no"
				if m.Read {
					read = "yes"
				}
				if m.Flagged {
					flagged = "yes"
				}
				if m.HasAttachments {
					att = "yes"
				}
				t.AddRow(m.Sender, m.Subject, m.Date, read, flagged, att)
			}
			t.Print()
		}
		return nil
	},
}

// --- read ---

var messagesReadCmd = &cobra.Command{
	Use:   "read <message-id>",
	Short: "Read full message content by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, err := mail.ReadMessage(args[0])
		if err != nil {
			output.PrintError("messages read", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("messages read", msg, prettyFlag)
			return nil
		}

		fmt.Printf("From:    %s\n", msg.Sender)
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("Date:    %s\n", msg.Date)
		fmt.Printf("Read:    %v  Flagged: %v\n", msg.Read, msg.Flagged)
		if len(msg.Attachments) > 0 {
			fmt.Printf("Attachments:\n")
			for _, a := range msg.Attachments {
				fmt.Printf("  - %s\n", a.Name)
			}
		}
		fmt.Println()
		fmt.Println(msg.Body)
		return nil
	},
}

// --- open ---

var messagesOpenCmd = &cobra.Command{
	Use:   "open <message-id>",
	Short: "Open a message in Mail.app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mail.OpenMessage(args[0]); err != nil {
			output.PrintError("messages open", "executionFailed", err.Error(), prettyFlag)
			return nil
		}
		if isJSON(cmd) {
			output.PrintJSON("messages open", map[string]string{"id": args[0], "status": "opened"}, prettyFlag)
			return nil
		}
		fmt.Println("Opened in Mail.app")
		return nil
	},
}

// --- search ---

var (
	msgSearchQuery   string
	msgSearchAccount string
	msgSearchMailbox string
	msgSearchFrom    string
	msgSearchSubject string
	msgSearchLimit   int
	msgSearchSummary bool
)

var messagesSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search messages",
	RunE: func(cmd *cobra.Command, args []string) error {
		msgs, err := mail.SearchMessages(mail.SearchMessagesParams{
			Query:       msgSearchQuery,
			AccountName: msgSearchAccount,
			MailboxName: msgSearchMailbox,
			From:        msgSearchFrom,
			Subject:     msgSearchSubject,
			Limit:       msgSearchLimit,
			IncludeBody: msgSearchSummary,
		})
		if err != nil {
			output.PrintError("messages search", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("messages search", msgs, prettyFlag)
			return nil
		}

		fmt.Printf("%d result(s)\n\n", len(msgs))
		if msgSearchSummary {
			printSummaryTable(msgs)
		} else {
			t := output.NewTable("SENDER", "SUBJECT", "DATE", "READ")
			for _, m := range msgs {
				read := "no"
				if m.Read {
					read = "yes"
				}
				t.AddRow(m.Sender, m.Subject, m.Date, read)
			}
			t.Print()
		}
		return nil
	},
}

// printSummaryTable prints messages as a two-line-per-message summary table.
func printSummaryTable(msgs []mail.Message) {
	const subjectWidth = 45
	const dateWidth = 28
	const summaryWidth = 72

	header := fmt.Sprintf("%-*s  %-*s  %s", subjectWidth, "SUBJECT", dateWidth, "DATE", "SUMMARY")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	indent := strings.Repeat(" ", subjectWidth+2+dateWidth+2)
	for _, m := range msgs {
		subj := truncate(m.Subject, subjectWidth)
		date := truncate(m.Date, dateWidth)
		line1, line2 := mail.Summarize(m.Body, summaryWidth)
		fmt.Printf("%-*s  %-*s  %s\n", subjectWidth, subj, dateWidth, date, line1)
		if line2 != "" {
			fmt.Printf("%s%s\n", indent, line2)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func init() {
	messagesListCmd.Flags().StringVar(&msgListAccount, "account", "", "account name (required)")
	messagesListCmd.Flags().StringVar(&msgListMailbox, "mailbox", "", "mailbox name (required)")
	messagesListCmd.Flags().IntVar(&msgListLimit, "limit", 25, "max messages to return (0 = unlimited)")
	messagesListCmd.Flags().BoolVar(&msgListUnread, "unread", false, "only show unread messages")
	messagesListCmd.Flags().BoolVar(&msgListSummary, "summary", false, "fetch and display a 2-line body preview per message")

	messagesSearchCmd.Flags().StringVar(&msgSearchQuery, "query", "", "full-text search query")
	messagesSearchCmd.Flags().StringVar(&msgSearchAccount, "account", "", "filter by account")
	messagesSearchCmd.Flags().StringVar(&msgSearchMailbox, "mailbox", "", "restrict search to one mailbox (skips noise mailboxes by default when omitted)")
	messagesSearchCmd.Flags().StringVar(&msgSearchFrom, "from", "", "filter by sender address")
	messagesSearchCmd.Flags().StringVar(&msgSearchSubject, "subject", "", "filter by subject")
	messagesSearchCmd.Flags().IntVar(&msgSearchLimit, "limit", 25, "max results (0 = unlimited)")
	messagesSearchCmd.Flags().BoolVar(&msgSearchSummary, "summary", false, "fetch and display a 2-line body preview per message")

	messagesCmd.AddCommand(messagesListCmd)
	messagesCmd.AddCommand(messagesReadCmd)
	messagesCmd.AddCommand(messagesOpenCmd)
	messagesCmd.AddCommand(messagesSearchCmd)
	rootCmd.AddCommand(messagesCmd)
}
