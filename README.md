# apple-mail-cli

A Claude Code plugin marketplace with two read-only macOS automation tools. Both target macOS 13+ Apple Silicon (arm64) only and communicate with Apple apps via Apple Events — no credentials, no network access, no API keys.

| Plugin | Binary | What it does |
|--------|--------|--------------|
| `apple-mail` | `amail` | Read and manage Mail.app — accounts, mailboxes, messages, rules, send |
| `ical` | `aical` | Read Calendar.app — calendars and events with date-range filtering |

**Platform:** macOS 13+ · Apple Silicon (arm64) only

## Install via Claude Code

Add this marketplace to Claude Code, then install the plugin(s) you want:

```sh
/plugin marketplace add trodemaster/apple-mail-cli
/plugin install apple-mail@amail-plugins
/plugin install ical@amail-plugins
```

Each plugin ships a pre-built, ad-hoc-signed arm64 binary in its `bin/` directory. The plugin system adds it to PATH automatically — no manual setup required. On first use, macOS will prompt for automation permission — click **Allow**.

## Build from source

Requires Go 1.21+ and macOS Apple Silicon.

```sh
make              # build, sign (ad-hoc), and install both binaries to $GOBIN
make skill        # cross-compile both arm64 plugin binaries and sign them
make skill-mail   # apple-mail plugin only
make skill-ical   # ical plugin only
```

## amail commands

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

## aical commands

### Calendars

```sh
aical calendars
```

Lists all calendars with id, name, writable flag, and description.

### Events

```sh
# Events for the next 7 days (default)
aical events

# Just today
aical events --today

# Specific date range
aical events --from 2026-05-01 --to 2026-05-31

# Relative window
aical events --days 14

# Filter to one calendar
aical events --calendar "Work" --days 30 [--limit 100]
```

Flags: `--calendar NAME`, `--from YYYY-MM-DD`, `--to YYYY-MM-DD`, `--today`, `--days N` (default 7), `--limit N` (default 50), `--format json`, `--pretty`

## Output

All commands (`amail` and `aical`) print a human-readable table by default. When piped or when `--format json` is passed, output is a JSON envelope:

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

Both binaries shell out to `osascript` and communicate with macOS apps via Apple Events. No network connections, no credential storage — they read the same data the apps show you.

Each binary is ad-hoc codesigned with the `com.apple.security.automation.apple-events` entitlement so macOS grants automation access without a Developer ID.

## Repository structure

```
apple-mail-cli/
├── .claude-plugin/
│   └── marketplace.json          # Root marketplace catalog (lists both plugins)
├── plugins/
│   ├── apple-mail/
│   │   ├── .claude-plugin/
│   │   │   └── plugin.json       # Plugin manifest
│   │   ├── bin/
│   │   │   └── amail             # Pre-built arm64 binary (auto-added to PATH)
│   │   └── skills/apple-mail/
│   │       └── SKILL.md          # Skill instructions
│   └── ical/
│       ├── .claude-plugin/
│       │   └── plugin.json
│       ├── bin/
│       │   └── aical             # Pre-built arm64 binary (auto-added to PATH)
│       └── skills/ical/
│           └── SKILL.md
├── cmd/                          # amail cobra commands
├── ical/                         # aical binary entrypoint + cobra commands
├── internal/
│   ├── mail/                     # AppleScript runners + typed ops for Mail.app
│   └── ical/                     # AppleScript runners + typed ops for Calendar.app
└── Makefile
```

To rebuild the bundled binaries after source changes:

```sh
make skill
```
