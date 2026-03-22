package mail

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed scripts/rules_list.applescript
var rulesListScript string

//go:embed scripts/rules_add_condition.applescript
var rulesAddConditionScript string

//go:embed scripts/rules_create.applescript
var rulesCreateScript string

//go:embed scripts/rules_apply_count.applescript
var rulesApplyCountScript string

//go:embed scripts/rules_apply_batch.applescript
var rulesApplyBatchScript string

//go:embed scripts/rules_apply_targeted.applescript
var rulesApplyTargetedScript string

// ruleTypeToAS maps short type names to AppleScript enum constants.
var ruleTypeToAS = map[string]string{
	"from":          "from header",
	"to":            "to header",
	"cc":            "cc header",
	"subject":       "subject header",
	"any-recipient": "any recipient",
	"body":          "message content",
	"account":       "account",
}

// qualifierToAS maps short qualifier names to AppleScript enum constants.
var qualifierToAS = map[string]string{
	"contains":     "does contain value",
	"not-contains": "does not contain value",
	"begins-with":  "begins with value",
	"ends-with":    "ends with value",
	"equals":       "equal to value",
}

type AddConditionParams struct {
	RuleName   string
	RuleType   string // short form: "from", "subject", etc.
	Qualifier  string // short form: "ends-with", "contains", etc.
	Expression string
}

type CreateRuleParams struct {
	RuleName    string
	MoveMailbox string
	AccountName string // optional; empty = search all accounts
	RuleType    string // short form
	Qualifier   string // short form
	Expression  string
}

type ruleConditionScriptParams struct {
	RuleName   string
	RuleTypeAS string
	QualifierAS string
	Expression  string
}

type ruleCreateScriptParams struct {
	RuleName    string
	MoveMailbox string
	AccountName string
	RuleTypeAS  string
	QualifierAS string
	Expression  string
}

func resolveRuleType(short string) (string, error) {
	if as, ok := ruleTypeToAS[short]; ok {
		return as, nil
	}
	// accept full AS string passthrough
	for _, v := range ruleTypeToAS {
		if v == short {
			return short, nil
		}
	}
	return "", fmt.Errorf("unknown rule type %q; valid: %s", short, strings.Join(ruleTypeKeys(), ", "))
}

func resolveQualifier(short string) (string, error) {
	if as, ok := qualifierToAS[short]; ok {
		return as, nil
	}
	for _, v := range qualifierToAS {
		if v == short {
			return short, nil
		}
	}
	return "", fmt.Errorf("unknown qualifier %q; valid: %s", short, strings.Join(qualifierKeys(), ", "))
}

func ruleTypeKeys() []string {
	keys := make([]string, 0, len(ruleTypeToAS))
	for k := range ruleTypeToAS {
		keys = append(keys, k)
	}
	return keys
}

func qualifierKeys() []string {
	keys := make([]string, 0, len(qualifierToAS))
	for k := range qualifierToAS {
		keys = append(keys, k)
	}
	return keys
}

func AddRuleCondition(p AddConditionParams) error {
	rtAS, err := resolveRuleType(p.RuleType)
	if err != nil {
		return err
	}
	qualAS, err := resolveQualifier(p.Qualifier)
	if err != nil {
		return err
	}
	_, err = RenderScript(rulesAddConditionScript, ruleConditionScriptParams{
		RuleName:    p.RuleName,
		RuleTypeAS:  rtAS,
		QualifierAS: qualAS,
		Expression:  p.Expression,
	})
	return err
}

type ApplyRulesParams struct {
	MailboxName string // default: "INBOX"
	AccountName string // optional; empty = all accounts
	BatchSize   int    // default: 500
}

type ApplyRulesResult struct {
	Account       string `json:"account"`
	Mailbox       string `json:"mailbox"`
	OriginalCount int    `json:"originalCount"`
	MovedCount    int    `json:"movedCount"`
	FinalCount    int    `json:"finalCount"`
}

type applyBatchParams struct {
	AccountName string
	MailboxName string
	BatchSize   int
}

type applyCountParams struct {
	AccountName string
	MailboxName string
}

// ApplyRules applies all enabled Mail.app rules to a mailbox by running
// batched osascript calls from Go, avoiding the 5-minute timeout.
// The progress callback (if non-nil) is called after each batch with
// (accountName, batchNum, totalBatches, movedSoFar).
func ApplyRules(p ApplyRulesParams, progress func(account string, batch, total, moved int)) ([]ApplyRulesResult, error) {
	if p.MailboxName == "" {
		p.MailboxName = "INBOX"
	}
	if p.BatchSize <= 0 {
		p.BatchSize = 500
	}

	// Get all account names first.
	accounts, err := ListAccounts()
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	var results []ApplyRulesResult
	for _, acc := range accounts {
		if p.AccountName != "" && acc.Name != p.AccountName {
			continue
		}

		// Get the message count for this account's mailbox.
		countOut, err := RenderScript(rulesApplyCountScript, applyCountParams{
			AccountName: acc.Name,
			MailboxName: p.MailboxName,
		})
		if err != nil {
			// Account doesn't have this mailbox — skip.
			continue
		}
		originalCount := 0
		fmt.Sscanf(strings.TrimSpace(countOut), "%d", &originalCount)
		if originalCount == 0 {
			continue
		}

		totalBatches := (originalCount + p.BatchSize - 1) / p.BatchSize
		totalMoved := 0

		for batch := 1; batch <= totalBatches; batch++ {
			if progress != nil {
				progress(acc.Name, batch, totalBatches, totalMoved)
			}
			out, err := RenderScriptResilient(rulesApplyBatchScript, applyBatchParams{
				AccountName: acc.Name,
				MailboxName: p.MailboxName,
				BatchSize:   p.BatchSize,
			}, nil)
			if err != nil {
				return nil, fmt.Errorf("apply batch %d/%d for %s: %w", batch, totalBatches, acc.Name, err)
			}
			moved := 0
			fmt.Sscanf(strings.TrimSpace(out), "%d", &moved)
			totalMoved += moved
		}

		// Get final count.
		finalOut, _ := RenderScript(rulesApplyCountScript, applyCountParams{
			AccountName: acc.Name,
			MailboxName: p.MailboxName,
		})
		finalCount := 0
		fmt.Sscanf(strings.TrimSpace(finalOut), "%d", &finalCount)

		results = append(results, ApplyRulesResult{
			Account:       acc.Name,
			Mailbox:       p.MailboxName,
			OriginalCount: originalCount,
			MovedCount:    totalMoved,
			FinalCount:    finalCount,
		})
	}
	return results, nil
}

type ApplyRuleTargetedParams struct {
	RuleName    string
	MailboxName string // source mailbox, default "INBOX"
	AccountName string // source account
}

type applyTargetedScriptParams struct {
	RuleName    string
	MailboxName string
	AccountName string
}

// ApplyRuleTargeted applies a single rule to a mailbox using indexed 'whose' queries
// per condition instead of perform mail action. Fast even on large mailboxes.
// If Mail.app hangs, it is automatically restarted and the operation retried.
func ApplyRuleTargeted(p ApplyRuleTargetedParams) (int, error) {
	if p.MailboxName == "" {
		p.MailboxName = "INBOX"
	}
	out, err := RenderScriptResilient(rulesApplyTargetedScript, applyTargetedScriptParams{
		RuleName:    p.RuleName,
		MailboxName: p.MailboxName,
		AccountName: p.AccountName,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("apply rule %q: %w", p.RuleName, err)
	}
	moved := 0
	fmt.Sscanf(strings.TrimSpace(out), "%d", &moved)
	return moved, nil
}

func CreateRule(p CreateRuleParams) error {
	rtAS, err := resolveRuleType(p.RuleType)
	if err != nil {
		return err
	}
	qualAS, err := resolveQualifier(p.Qualifier)
	if err != nil {
		return err
	}
	_, err = RenderScript(rulesCreateScript, ruleCreateScriptParams{
		RuleName:    p.RuleName,
		MoveMailbox: p.MoveMailbox,
		AccountName: p.AccountName,
		RuleTypeAS:  rtAS,
		QualifierAS: qualAS,
		Expression:  p.Expression,
	})
	return err
}

type RuleCondition struct {
	RuleType   string `json:"ruleType"`
	Qualifier  string `json:"qualifier"`
	Expression string `json:"expression"`
	Header     string `json:"header,omitempty"`
}

type RuleActions struct {
	MoveToMailbox string `json:"moveToMailbox,omitempty"`
	CopyToMailbox string `json:"copyToMailbox,omitempty"`
	DeleteMessage bool   `json:"deleteMessage,omitempty"`
	MarkRead      bool   `json:"markRead,omitempty"`
	MarkFlagged   bool   `json:"markFlagged,omitempty"`
	ForwardTo     string `json:"forwardTo,omitempty"`
	RedirectTo    string `json:"redirectTo,omitempty"`
}

type Rule struct {
	Name                   string          `json:"name"`
	Enabled                bool            `json:"enabled"`
	AllConditionsMustBeMet bool            `json:"allConditionsMustBeMet"`
	StopEvaluatingRules    bool            `json:"stopEvaluatingRules"`
	Conditions             []RuleCondition `json:"conditions"`
	Actions                RuleActions     `json:"actions"`
}

func ListRules() ([]Rule, error) {
	out, err := RunScript(rulesListScript)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	var rules []Rule
	if err := json.Unmarshal([]byte(out), &rules); err != nil {
		return nil, fmt.Errorf("parse rules: %w (raw: %s)", err, out)
	}
	return rules, nil
}
