---
name: apple-mail
description: "Read and manage Apple Mail.app on macOS Apple Silicon. Use when the user asks about their email, wants to read or find messages, check unread mail, list accounts or mailboxes, send an email, route a domain to a folder, create or apply filter rules, or anything related to Mail.app or email on macOS. Requires: macOS on Apple Silicon (arm64)."
argument-hint: "email task — e.g. 'show unread', 'find emails from X', 'send to Y'"
allowed-tools: Bash(amail *)
---

# apple-mail

Interact with Apple Mail.app via the bundled `amail` CLI (Apple Silicon arm64 binary, no credentials, no IMAP, no API keys). The plugin system automatically adds `amail` to PATH — no manual setup required.

## Requirements

- macOS 13+ on Apple Silicon (arm64)
- Mail.app open and running
- Plugin installed via the Claude Code marketplace

On first `amail` invocation, macOS will prompt for Mail.app automation permission — tell the user to click **Allow**.

Use the `amail` CLI to interact with Apple Mail.app via Apple Events. Mail.app must be running. On first use macOS will prompt once for automation permission — tell the user to click Allow.

## Discover available accounts and mailboxes

List all configured accounts:
```
amail accounts
```

List all mailboxes (optionally filtered to one account):
```
amail mailboxes [--account "Account Name"]
```

**Important:** The user may have mailboxes named after people (e.g. `chandra`, `mom`, `work`). Always check `amail mailboxes` before searching — if a person-named mailbox exists, target it directly with `messages list --mailbox` instead of using `messages search`. This is orders of magnitude faster.

## List messages

List messages in a specific mailbox (account and mailbox are required):
```
amail messages list --account "Account Name" --mailbox "INBOX" \
  [--limit N] [--unread] [--summary]
```

- `--limit` defaults to 25; use `--limit 0` for all messages
- `--unread` restricts to unread messages only
- `--summary` fetches message bodies and displays a 2-line plain-text preview per message (see below)

Output fields per message: `id`, `subject`, `sender`, `date`, `read`, `flagged`, `hasAttachments`, `body` (populated when `--summary` is used)

## "Last N emails from a person" — the preferred pattern

When the user asks for recent emails from someone (e.g. "last 10 emails from chandra"):

1. Run `amail mailboxes` to check if a person-named mailbox exists
2. If yes — use `messages list` targeting that mailbox directly:
   ```
   amail messages list --account "iCloud" --mailbox "chandra" --limit 10 --summary
   ```
3. If no — use `messages search` with `--from`:
   ```
   amail messages search --from "chandra" --limit 10 --summary
   ```

The `--summary` flag produces a table with subject, date, and a 2-line body preview. Body processing (HTML stripping, whitespace normalization, word-wrap) is done in Go — not in the shell.

## Read a full message

```
amail messages read <message-id>
```

Returns: `id`, `subject`, `sender`, `date`, `read`, `flagged`, `body`, `attachments[]`

## Open a message in Mail.app

```
amail messages open <message-id>
```

Opens the message in a Mail.app viewer window and brings Mail.app to the front. Use this when the user wants to view or act on a specific message in the Mail.app GUI rather than reading it in the terminal.

## Search messages

```
amail messages search [--query "text"] [--account "Account Name"] \
  [--mailbox "Mailbox Name"] [--from "addr"] [--subject "text"] \
  [--limit N] [--summary]
```

- `--from` and `--subject` use Mail.app's indexed `whose` clause — fast even on large mailboxes
- `--query` searches subject, sender, and body — slower; always add `--mailbox` to scope it
- `--mailbox` restricts to one mailbox; without it, Trash/Junk/Sent/Drafts/All Mail are automatically skipped
- `--limit` defaults to 25; use `--limit 0` for all results
- `--summary` fetches and displays a 2-line body preview per result

## Send email

```
amail send --to "addr@example.com" --subject "Subject" --body "Body text" \
  [--to "addr2@example.com"] \
  [--cc "addr@example.com"] \
  [--bcc "addr@example.com"] \
  [--attachment "/absolute/path/to/file"]
```

- `--to`, `--cc`, `--bcc`, and `--attachment` are all repeatable
- Attachment paths must be absolute

## JSON output

All commands emit a JSON envelope when piped or when `--format json` is passed:

```json
{
  "ok": true,
  "data": <T>,
  "error": { "code": "...", "message": "..." } | null,
  "meta": { "command": "...", "timestamp": "..." }
}
```

Use `--format json` whenever you need to parse output. Use `--pretty` for human-readable JSON. Check `ok` before reading `data`.

```
amail messages list --account "iCloud" --mailbox "chandra" --limit 10 --format json
amail messages search --from "chandra" --limit 10 --format json | jq '.data[].subject'
```

## Full command/flag reference

```
amail schema
```

Emits the complete JSON spec of every command, subcommand, flag, type, default, and description. Call this first if you are unsure of available options.

## Inspect and manage rules

List all Mail.app filter rules:
```
amail rules list
amail rules list --format json --pretty
```

Each rule has a name, enabled flag, match mode (`any` or `all`), a list of conditions, and an action. Conditions typically look like `from ends-with @domain.com`. Actions are typically `move: FolderName`.

### Add a condition to an existing rule

```
amail rules add-condition --rule "RULE_NAME" --expression "domain.com" \
  [--type from] [--qualifier ends-with]
```

- `--type` default `from`; options: `from`, `to`, `cc`, `subject`, `any-recipient`, `body`, `account`
- `--qualifier` default `ends-with`; options: `contains`, `not-contains`, `begins-with`, `ends-with`, `equals`

### Create a new rule

```
amail rules create --name "services2" --move "Services" --expression "domain.com" \
  [--account "Account Name"] [--type from] [--qualifier ends-with]
```

Creates a new enabled rule with OR logic and an initial condition. `--account` is optional — omit to search all accounts for the destination mailbox.

### Apply rules to a mailbox

**Targeted mode (fast — recommended):** applies a single named rule using indexed `whose` queries:
```
amail rules apply --rule "services" [--mailbox INBOX] [--account "Account Name"]
```

**General mode (slower):** applies all enabled rules using batched `perform mail action`:
```
amail rules apply [--mailbox INBOX] [--account "Account Name"] [--batch-size 500]
```

Targeted mode is orders of magnitude faster on large mailboxes. Always prefer `--rule` when you know which rule to apply.

### Routing a domain to a folder — full workflow

The user's pattern: one rule per destination folder, OR logic (`any`), `from ends-with @domain.com` conditions.

1. Find existing rules for the target folder and their condition counts:
   ```
   amail rules list --format json | jq '.data[] | select(.actions.moveToMailbox == "Services") | {name, count: (.conditions | length)}'
   ```
2. If a rule has **< 80 conditions** → add to it with `rules add-condition`
3. If a rule has **≥ 80 conditions** → create a new sibling with `rules create`:
   - `services` full → `services2`; `services2` full → `services3`
4. Apply the rule to move existing INBOX messages:
   ```
   amail rules apply --rule "services"
   ```

**Why 80?** Mail.app's GUI degrades noticeably beyond 80 conditions per rule.

## Guidelines

- Always run `amail accounts` first if you don't know the account name — never guess it
- Run `amail mailboxes` to check for person-named mailboxes before doing a sender search — direct mailbox targeting is far faster than search
- `--from` and `--subject` on `messages search` are indexed and fast; `--query` (body search) is slow without `--mailbox` scoping
- Use `--format json` whenever you need to extract data (e.g. a message `id` to pass to `read`)
- Message `id` values come from `messages list` or `messages search` output — they are opaque strings, never construct them manually
- Before sending, show the user a plain-English summary (to, subject, body preview) and confirm
- If `ok` is `false` in JSON output, surface the `error.message` to the user and stop
- Attachment paths passed to `--attachment` must be absolute; expand `~` before passing
- Mail.app must be open and running; if an osascript error occurs, ask the user to open Mail.app and try again
- Each rule must have ≤ 80 conditions — create a numbered sibling rule when an existing rule is full
