package mail

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed scripts/accounts_list.applescript
var accountsListScript string

type Account struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	EmailAddresses []string `json:"emailAddresses"`
	Type           string   `json:"type"`
}

func ListAccounts() ([]Account, error) {
	out, err := RunScript(accountsListScript)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	var accounts []Account
	if err := json.Unmarshal([]byte(out), &accounts); err != nil {
		return nil, fmt.Errorf("parse accounts: %w (raw: %s)", err, out)
	}
	return accounts, nil
}
