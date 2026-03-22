# amail

A command-line interface for Apple Mail.app on macOS. Reads and manages your email — accounts, mailboxes, messages, and filter rules — via Apple Events (no credentials, no IMAP, no API keys).

## Requirements

- macOS with Mail.app configured
- Go 1.21+ (to build)
- On first run, macOS will prompt for automation permission — click **Allow**

## Install

```sh
make        # build, sign (ad-hoc), and install to $GOBIN
```

## Commands

### Accounts

```sh
amail accounts
```

Lists all configured email accounts with name, addresses, and type.

### Mailboxes

```sh
amail mailboxes [--account "Account Name"]
```

Lists all mailboxes with unread and total message counts. Filter to one account with `--account`.

### Messages

```sh
# List messages in a mailbox
amail messages list --account "iCloud" --mailbox "INBOX" [--limit 25] [--unread] [--summary]

# Read a full message by ID
amail messages read <message-id>

# Search across mailboxes
amail messages search [--from "addr"] [--subject "text"] [--query "text"] \
  [--account "Name"] [--mailbox "Name"] [--limit 25] [--summary]
```

- `--summary` fetches message bodies and shows a 2-line plain-text preview (HTML stripped, word-wrapped in Go)
- `--from` and `--subject` use Mail.app's indexed search — fast even on mailboxes with thousands of messages
- `--query` searches subject, sender, and body — slower; use `--mailbox` to scope it
- Without `--mailbox`, search automatically skips Trash, Junk, Sent, Drafts, and All Mail

### Send

```sh
amail send --to "addr@example.com" --subject "Subject" --body "Body text" \
  [--to "addr2@example.com"] \
  [--cc "addr@example.com"] \
  [--bcc "addr@example.com"] \
  [--attachment "/absolute/path/to/file"]
```

`--to`, `--cc`, `--bcc`, and `--attachment` are all repeatable.

### Rules

```sh
# List all filter rules with conditions and actions
amail rules list

# Add a condition to an existing rule
amail rules add-condition --rule "services" --expression "example.com" \
  [--type from] [--qualifier ends-with]

# Create a new rule that moves matching messages to a mailbox
amail rules create --name "services2" --move "Services" --expression "example.com" \
  [--account "Account Name"] [--type from] [--qualifier ends-with]

# Apply a single rule to a mailbox using fast indexed queries (recommended)
amail rules apply --rule "services" [--mailbox INBOX] [--account "Account Name"]

# Apply all enabled rules (batched; slower on large mailboxes)
amail rules apply [--mailbox INBOX] [--batch-size 500]
```

Condition types: `from` `to` `cc` `subject` `any-recipient` `body` `account`
Qualifiers: `contains` `not-contains` `begins-with` `ends-with` `equals`

**Rule scale:** Mail.app's GUI degrades beyond ~80 conditions per rule. When a rule is full, create a numbered sibling (`services2`, `services3`, etc.).

### Schema

```sh
amail schema
```

Emits the full JSON spec of every command, subcommand, flag, type, and default. Useful for agents and scripting.

## Output

All commands print a human-readable table by default. When piped or when `--format json` is passed, output is a JSON envelope:

```json
{
  "ok": true,
  "data": <T>,
  "error": { "code": "...", "message": "..." } | null,
  "meta": { "command": "...", "timestamp": "..." }
}
```

Use `--pretty` for formatted JSON.

```sh
amail messages search --from "someone@example.com" --limit 10 --format json | jq '.data[].subject'
amail rules list --format json | jq '.data[] | select(.actions.moveToMailbox == "Work") | {name, count: (.conditions | length)}'
```

## How it works

`amail` shells out to `osascript` and communicates with Mail.app via Apple Events. No network connections, no credential storage — it reads and writes the same data Mail.app shows you.

The binary is ad-hoc codesigned with the `com.apple.security.automation.apple-events` entitlement so macOS grants it automation access.

## Claude skill

A Claude Code skill lives at `.claude/skills/apple-mail/SKILL.md`. It gives Claude full coverage of all commands and the domain-routing workflow, optimised for agent use.

AppleScript + Go compatibility notes (quoting rules, known Mail.app scripting bugs, performance patterns) are documented in `.claude/skills/apple-mail/GO-APPLESCRIPT-NOTES.md`.
