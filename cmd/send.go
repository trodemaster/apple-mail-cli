package cmd

import (
	"fmt"

	"github.com/blakemcanally/apple-mail-cli/internal/mail"
	"github.com/blakemcanally/apple-mail-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	sendTo          []string
	sendCC          []string
	sendBCC         []string
	sendSubject     string
	sendBody        string
	sendAttachments []string
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Compose and send an email via Mail.app",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(sendTo) == 0 {
			return fmt.Errorf("--to is required")
		}
		if sendSubject == "" {
			return fmt.Errorf("--subject is required")
		}

		err := mail.Send(mail.SendParams{
			To:          sendTo,
			CC:          sendCC,
			BCC:         sendBCC,
			Subject:     sendSubject,
			Body:        sendBody,
			Attachments: sendAttachments,
		})
		if err != nil {
			output.PrintError("send", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("send", map[string]bool{"sent": true}, prettyFlag)
			return nil
		}

		fmt.Println("Message sent.")
		return nil
	},
}

func init() {
	sendCmd.Flags().StringArrayVar(&sendTo, "to", nil, "recipient address (repeatable)")
	sendCmd.Flags().StringArrayVar(&sendCC, "cc", nil, "CC address (repeatable)")
	sendCmd.Flags().StringArrayVar(&sendBCC, "bcc", nil, "BCC address (repeatable)")
	sendCmd.Flags().StringVar(&sendSubject, "subject", "", "email subject (required)")
	sendCmd.Flags().StringVar(&sendBody, "body", "", "email body")
	sendCmd.Flags().StringArrayVar(&sendAttachments, "attachment", nil, "file path to attach (repeatable)")
	rootCmd.AddCommand(sendCmd)
}
