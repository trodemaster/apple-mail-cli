package mail

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed scripts/mailboxes_list.applescript
var mailboxesListScript string

type Mailbox struct {
	Name        string `json:"name"`
	Account     string `json:"account"`
	UnreadCount int    `json:"unreadCount"`
	TotalCount  int    `json:"totalCount"`
}

type mailboxesListParams struct {
	AccountName string
}

func ListMailboxes(accountName string) ([]Mailbox, error) {
	out, err := RenderScript(mailboxesListScript, mailboxesListParams{AccountName: accountName})
	if err != nil {
		return nil, fmt.Errorf("list mailboxes: %w", err)
	}
	var mailboxes []Mailbox
	if err := json.Unmarshal([]byte(out), &mailboxes); err != nil {
		return nil, fmt.Errorf("parse mailboxes: %w (raw: %s)", err, out)
	}
	return mailboxes, nil
}
