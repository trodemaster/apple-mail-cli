package cmd

import (
	"fmt"
	"strings"

	"github.com/trodemaster/apple-mail-cli/internal/mail"
	"github.com/trodemaster/apple-mail-cli/internal/output"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Rule subcommands: list",
}

var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Mail.app rules with conditions and actions",
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := mail.ListRules()
		if err != nil {
			output.PrintError("rules list", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("rules list", rules, prettyFlag)
			return nil
		}

		fmt.Printf("%d rule(s)\n\n", len(rules))

		t := output.NewTable("NAME", "EN", "MATCH", "CONDITION(S)", "ACTION")
		for _, r := range rules {
			enabled := "no"
			if r.Enabled {
				enabled = "yes"
			}
			match := "any"
			if r.AllConditionsMustBeMet {
				match = "all"
			}
			action := formatRuleAction(r.Actions)

			if len(r.Conditions) == 0 {
				t.AddRow(r.Name, enabled, match, "(no conditions)", action)
				continue
			}
			for i, cond := range r.Conditions {
				if i == 0 {
					t.AddRow(r.Name, enabled, match, formatRuleCondition(cond), action)
				} else {
					t.AddRow("", "", "", formatRuleCondition(cond), "")
				}
			}
		}
		t.Print()
		return nil
	},
}

func formatRuleCondition(c mail.RuleCondition) string {
	ruleType := shortRuleType(c.RuleType)
	if c.Header != "" {
		ruleType = "header:" + c.Header
	}
	return ruleType + " " + shortQualifier(c.Qualifier) + " " + c.Expression
}

func formatRuleAction(a mail.RuleActions) string {
	var parts []string
	if a.MoveToMailbox != "" {
		parts = append(parts, "move: "+a.MoveToMailbox)
	}
	if a.CopyToMailbox != "" {
		parts = append(parts, "copy: "+a.CopyToMailbox)
	}
	if a.DeleteMessage {
		parts = append(parts, "delete")
	}
	if a.MarkRead {
		parts = append(parts, "mark-read")
	}
	if a.MarkFlagged {
		parts = append(parts, "mark-flagged")
	}
	if a.ForwardTo != "" {
		parts = append(parts, "forward: "+a.ForwardTo)
	}
	if a.RedirectTo != "" {
		parts = append(parts, "redirect: "+a.RedirectTo)
	}
	if len(parts) == 0 {
		return "(no action)"
	}
	return strings.Join(parts, ", ")
}

func shortRuleType(s string) string {
	switch s {
	case "from header":
		return "from"
	case "to header":
		return "to"
	case "subject header":
		return "subject"
	case "any recipient":
		return "any-recipient"
	case "message content":
		return "body"
	case "account":
		return "account"
	case "cc header":
		return "cc"
	case "header key":
		return "header"
	default:
		return s
	}
}

func shortQualifier(s string) string {
	switch s {
	case "does contain value":
		return "contains"
	case "does not contain value":
		return "not-contains"
	case "begins with value":
		return "begins-with"
	case "ends with value":
		return "ends-with"
	case "equal to value":
		return "equals"
	default:
		return s
	}
}

var rulesApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply rules to a mailbox. Use --rule for fast targeted mode; omit for all rules (slower).",
	RunE: func(cmd *cobra.Command, args []string) error {
		mailboxName, _ := cmd.Flags().GetString("mailbox")
		accountName, _ := cmd.Flags().GetString("account")
		ruleName, _ := cmd.Flags().GetString("rule")
		batchSize, _ := cmd.Flags().GetInt("batch-size")

		stderr := cmd.ErrOrStderr()

		// Targeted mode: apply a single rule using indexed 'whose' queries — fast.
		if ruleName != "" {
			fmt.Fprintf(stderr, "Applying rule %q to %q (targeted)...\n", ruleName, mailboxName)

			// If no account specified, iterate all accounts.
			accounts := []string{accountName}
			if accountName == "" {
				accs, err := mail.ListAccounts()
				if err != nil {
					output.PrintError("rules apply", "executionFailed", err.Error(), prettyFlag)
					return nil
				}
				accounts = make([]string, len(accs))
				for i, a := range accs {
					accounts[i] = a.Name
				}
			}

			type targetedResult struct {
				Account string `json:"account"`
				Mailbox string `json:"mailbox"`
				Moved   int    `json:"moved"`
			}
			var results []targetedResult
			totalMoved := 0
			for _, acc := range accounts {
				moved, err := mail.ApplyRuleTargeted(mail.ApplyRuleTargetedParams{
					RuleName:    ruleName,
					MailboxName: mailboxName,
					AccountName: acc,
				})
				if err != nil {
					// Rule may not apply to this account's mailbox — skip.
					continue
				}
				results = append(results, targetedResult{Account: acc, Mailbox: mailboxName, Moved: moved})
				totalMoved += moved
			}

			if isJSON(cmd) {
				output.PrintJSON("rules apply", results, prettyFlag)
				return nil
			}
			fmt.Printf("%d message(s) moved\n", totalMoved)
			return nil
		}

		// General mode: apply all rules using batched perform mail action.
		fmt.Fprintf(stderr, "Applying all rules to %q (batched — may take a while for large mailboxes)...\n", mailboxName)

		results, err := mail.ApplyRules(mail.ApplyRulesParams{
			MailboxName: mailboxName,
			AccountName: accountName,
			BatchSize:   batchSize,
		}, func(account string, batch, total, moved int) {
			fmt.Fprintf(stderr, "  %s: batch %d/%d (%d moved so far)\n", account, batch, total, moved)
		})
		if err != nil {
			output.PrintError("rules apply", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("rules apply", results, prettyFlag)
			return nil
		}

		totalMoved := 0
		for _, r := range results {
			totalMoved += r.MovedCount
		}
		fmt.Printf("%d message(s) moved across %d account(s)\n\n", totalMoved, len(results))

		t := output.NewTable("ACCOUNT", "MAILBOX", "BEFORE", "MOVED", "AFTER")
		for _, r := range results {
			t.AddRow(r.Account, r.Mailbox,
				fmt.Sprintf("%d", r.OriginalCount),
				fmt.Sprintf("%d", r.MovedCount),
				fmt.Sprintf("%d", r.FinalCount))
		}
		t.Print()
		return nil
	},
}

var rulesAddConditionCmd = &cobra.Command{
	Use:   "add-condition",
	Short: "Add a condition to an existing rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		ruleName, _ := cmd.Flags().GetString("rule")
		ruleType, _ := cmd.Flags().GetString("type")
		qualifier, _ := cmd.Flags().GetString("qualifier")
		expression, _ := cmd.Flags().GetString("expression")

		if err := mail.AddRuleCondition(mail.AddConditionParams{
			RuleName:   ruleName,
			RuleType:   ruleType,
			Qualifier:  qualifier,
			Expression: expression,
		}); err != nil {
			output.PrintError("rules add-condition", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("rules add-condition", map[string]string{"rule": ruleName, "expression": expression}, prettyFlag)
			return nil
		}
		fmt.Printf("Added condition: %s %s %s → rule %q\n", ruleType, qualifier, expression, ruleName)
		return nil
	},
}

var rulesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new rule that moves matching messages to a mailbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		move, _ := cmd.Flags().GetString("move")
		account, _ := cmd.Flags().GetString("account")
		ruleType, _ := cmd.Flags().GetString("type")
		qualifier, _ := cmd.Flags().GetString("qualifier")
		expression, _ := cmd.Flags().GetString("expression")

		if err := mail.CreateRule(mail.CreateRuleParams{
			RuleName:    name,
			MoveMailbox: move,
			AccountName: account,
			RuleType:    ruleType,
			Qualifier:   qualifier,
			Expression:  expression,
		}); err != nil {
			output.PrintError("rules create", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("rules create", map[string]string{"name": name, "move": move, "expression": expression}, prettyFlag)
			return nil
		}
		fmt.Printf("Created rule %q → move to %q (first condition: %s %s %s)\n", name, move, ruleType, qualifier, expression)
		return nil
	},
}

func init() {
	rulesApplyCmd.Flags().String("mailbox", "INBOX", "Mailbox to apply rules to")
	rulesApplyCmd.Flags().String("account", "", "Account name (optional; applies to all accounts if omitted)")
	rulesApplyCmd.Flags().String("rule", "", "Apply only this rule using fast indexed queries (recommended for large mailboxes)")
	rulesApplyCmd.Flags().Int("batch-size", 500, "Messages per batch (general mode only)")

	rulesAddConditionCmd.Flags().String("rule", "", "Rule name to add the condition to (required)")
	rulesAddConditionCmd.Flags().String("type", "from", "Condition type: from, to, cc, subject, any-recipient, body, account")
	rulesAddConditionCmd.Flags().String("qualifier", "ends-with", "Qualifier: contains, not-contains, begins-with, ends-with, equals")
	rulesAddConditionCmd.Flags().String("expression", "", "Value to match (required)")
	_ = rulesAddConditionCmd.MarkFlagRequired("rule")
	_ = rulesAddConditionCmd.MarkFlagRequired("expression")

	rulesCreateCmd.Flags().String("name", "", "Name for the new rule (required)")
	rulesCreateCmd.Flags().String("move", "", "Destination mailbox name (required)")
	rulesCreateCmd.Flags().String("account", "", "Account containing the mailbox (optional; searches all if omitted)")
	rulesCreateCmd.Flags().String("type", "from", "Condition type: from, to, cc, subject, any-recipient, body, account")
	rulesCreateCmd.Flags().String("qualifier", "ends-with", "Qualifier: contains, not-contains, begins-with, ends-with, equals")
	rulesCreateCmd.Flags().String("expression", "", "Value to match (required)")
	_ = rulesCreateCmd.MarkFlagRequired("name")
	_ = rulesCreateCmd.MarkFlagRequired("move")
	_ = rulesCreateCmd.MarkFlagRequired("expression")

	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesApplyCmd)
	rulesCmd.AddCommand(rulesAddConditionCmd)
	rulesCmd.AddCommand(rulesCreateCmd)
	rootCmd.AddCommand(rulesCmd)
}
