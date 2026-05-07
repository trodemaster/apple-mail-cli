package mail

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

//go:embed scripts/messages_list.applescript
var messagesListScript string

//go:embed scripts/messages_read.applescript
var messagesReadScript string

//go:embed scripts/messages_search.applescript
var messagesSearchScript string

type Message struct {
	ID              string       `json:"id"`
	Subject         string       `json:"subject"`
	Sender          string       `json:"sender"`
	Date            string       `json:"date"`
	Read            bool         `json:"read"`
	Flagged         bool         `json:"flagged"`
	HasAttachments  bool         `json:"hasAttachments"`
	Body            string       `json:"body,omitempty"`
	Attachments     []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Name string `json:"name"`
}

type ListMessagesParams struct {
	AccountName string
	MailboxName string
	Limit       int
	UnreadOnly  bool
	IncludeBody bool
}

type SearchMessagesParams struct {
	Query        string
	AccountName  string
	From         string
	Subject      string
	Limit        int
	IncludeBody  bool
	// MailboxName restricts the search to a single mailbox (optional).
	MailboxName  string
}

func ListMessages(p ListMessagesParams) ([]Message, error) {
	// AppleScript booleans need lowercase true/false strings
	type scriptParams struct {
		AccountName string
		MailboxName string
		Limit       int
		UnreadOnly  string
		IncludeBody bool
	}
	unreadOnly := "false"
	if p.UnreadOnly {
		unreadOnly = "true"
	}
	out, err := RenderScript(messagesListScript, scriptParams{
		AccountName: p.AccountName,
		MailboxName: p.MailboxName,
		Limit:       p.Limit,
		UnreadOnly:  unreadOnly,
		IncludeBody: p.IncludeBody,
	})
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	var msgs []Message
	if err := json.Unmarshal([]byte(out), &msgs); err != nil {
		return nil, fmt.Errorf("parse messages: %w (raw: %s)", err, out)
	}
	return msgs, nil
}

func ReadMessage(messageID string) (*Message, error) {
	type scriptParams struct {
		MessageID string
	}
	out, err := RenderScript(messagesReadScript, scriptParams{MessageID: messageID})
	if err != nil {
		return nil, fmt.Errorf("read message: %w", err)
	}
	var msg Message
	if err := json.Unmarshal([]byte(out), &msg); err != nil {
		return nil, fmt.Errorf("parse message: %w (raw: %s)", err, out)
	}
	return &msg, nil
}

func OpenMessage(messageID string) error {
	// Use the message: URL scheme for a direct indexed lookup instead of
	// scanning all mailboxes via AppleScript, which hangs on large accounts.
	url := "message://%3C" + messageID + "%3E"
	out, err := exec.Command("open", url).CombinedOutput()
	if err != nil {
		return fmt.Errorf("open message: %s: %w", strings.TrimSpace(string(out)), err)
	}
	// Bring Mail.app to the front.
	_, _ = exec.Command("osascript", "-e", `tell application "Mail" to activate`).Output()
	return nil
}

func SearchMessages(p SearchMessagesParams) ([]Message, error) {
	out, err := RenderScript(messagesSearchScript, p)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	var msgs []Message
	if err := json.Unmarshal([]byte(out), &msgs); err != nil {
		return nil, fmt.Errorf("parse search results: %w (raw: %s)", err, out)
	}
	return msgs, nil
}
