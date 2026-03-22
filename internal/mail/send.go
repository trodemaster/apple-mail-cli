package mail

import (
	_ "embed"
	"fmt"
)

//go:embed scripts/send.applescript
var sendScript string

type SendParams struct {
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	Attachments []string
}

func Send(p SendParams) error {
	_, err := RenderScript(sendScript, p)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	return nil
}
