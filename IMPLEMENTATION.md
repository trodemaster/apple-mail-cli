# Implementation Reference

State as of 2026-05-06. Intended for agents updating or extending this repo.

## Purpose

Two standalone CLI binaries (`amail`, `aical`) that automate Apple Mail.app and Calendar.app on macOS Apple Silicon via Apple Events. No credentials, no network access, no API keys — they talk to the apps the user already has running. Packaged as a Claude Code plugin marketplace so Claude agents can invoke them as skills.

## Repository Layout

```
apple-mail-cli/
├── main.go                          # amail entrypoint → cmd.Execute()
├── ical/main.go                     # aical entrypoint → ical/cmd.Execute()
├── cmd/                             # amail cobra commands
│   ├── root.go                      # global --format / --pretty flags, isJSON()
│   ├── accounts.go
│   ├── mailboxes.go
│   ├── messages.go
│   ├── send.go
│   ├── rules.go
│   └── schema.go
├── ical/cmd/                        # aical cobra commands
│   ├── root.go                      # same flag pattern as cmd/root.go
│   ├── calendars.go
│   └── events.go
├── internal/
│   ├── mail/                        # Mail.app business logic
│   │   ├── client.go                # RunScript, RunScriptResilient, RenderScript*
│   │   ├── accounts.go
│   │   ├── mailboxes.go
│   │   ├── messages.go
│   │   ├── send.go
│   │   ├── rules.go
│   │   ├── summarize.go             # HTML → 2-line plain-text preview
│   │   └── scripts/                 # AppleScript templates (go:embed)
│   ├── ical/                        # Calendar.app business logic
│   │   ├── client.go                # RunScript, RenderScript (no resilience layer)
│   │   ├── calendars.go
│   │   ├── events.go
│   │   └── scripts/
│   └── output/
│       ├── envelope.go              # JSON envelope: {ok, data, error, meta}
│       └── table.go                 # auto-width tabular printer
├── plugins/
│   ├── apple-mail/
│   │   ├── .claude-plugin/plugin.json
│   │   ├── bin/amail                # pre-built arm64 binary (tracked in git)
│   │   ├── bin/entitlements.plist
│   │   └── skills/apple-mail/SKILL.md
│   └── ical/
│       ├── .claude-plugin/plugin.json
│       ├── bin/aical                # pre-built arm64 binary (tracked in git)
│       ├── bin/entitlements.plist
│       └── skills/ical/SKILL.md
├── .claude-plugin/
│   └── marketplace.json             # marketplace catalog (name: "amail-plugins")
├── entitlements.plist               # com.apple.security.automation.apple-events
├── Makefile
└── go.mod                           # module: github.com/trodemaster/apple-mail-cli
```

## Go Module

```
module github.com/trodemaster/apple-mail-cli
go 1.26.1

dependencies:
  github.com/spf13/cobra v1.10.2
  github.com/spf13/pflag v1.0.9
  golang.org/x/term v0.41.0   (TTY detection for auto-JSON mode)
  golang.org/x/sys v0.42.0
```

Both binaries live in the same module. `amail` builds from `main.go` (package `main`, imports `cmd`). `aical` builds from `ical/main.go` (package `main`, imports `ical/cmd`). Shared logic is in `internal/`.

## AppleScript Integration Pattern

All Apple Events calls go through `osascript -e <script>`. Scripts are stored as `.applescript` files in `internal/mail/scripts/` and `internal/ical/scripts/`, embedded at compile time with `//go:embed`, and templated with `text/template` before execution.

**Template rendering (both packages):**
```
RenderScript(scriptTmpl string, data interface{}) (string, error)
```
Renders the template with `data`, then calls `RunScript` (which runs `osascript -e`).

**Mail.app resilience layer** (`internal/mail/client.go`):
- `RunScript` — 5-minute timeout, no retry
- `RunScriptResilient` — 120-second timeout; detects Mail.app hangs (`AppleEvent timed out`, `signal: killed`, `Connection is invalid`); kills and relaunches Mail.app, retries once
- `RenderScriptResilient` — template rendering + resilience; optional `onRestart func()` callback for progress reporting
- `RestartMail()` — `pkill -x Mail`, waits 2s, `open -a Mail`, polls up to 15s for Mail.app to respond

The ical client (`internal/ical/client.go`) uses a simpler 2-minute timeout with no restart logic (Calendar.app is more stable).

**JSON serialization in AppleScript:** Scripts build JSON strings manually using an `escapeJSON` handler that escapes `"`, `\`, tab, LF, CR. AppleScript has no native JSON — this is intentional and correct.

## Output System

**JSON envelope** (`internal/output/envelope.go`):
```json
{ "ok": true, "data": <T>, "error": null, "meta": { "command": "...", "timestamp": "..." } }
{ "ok": false, "data": null, "error": { "code": "...", "message": "..." }, "meta": { ... } }
```
`PrintError` calls `os.Exit(1)` after printing the envelope. `PrintJSON` does not exit.

**Auto-JSON mode:** `isJSON(cmd)` in both `cmd/root.go` and `ical/cmd/root.go` returns true if `--format json` is set OR stdout is not a TTY (`golang.org/x/term`). This means piped output is always JSON.

**Table output** (`internal/output/table.go`): `NewTable(headers...)` → `AddRow(cols...)` → `Print()`. Auto-calculates column widths from content.

**Summary mode** (`internal/mail/summarize.go`): `Summarize(body, lineWidth)` strips HTML tags and entities, collapses whitespace, word-wraps to two lines. Used by `messages list --summary` and `messages search --summary`.

## amail Command Structure

```
amail [--format json] [--pretty]
  accounts                          list all configured accounts
  mailboxes [--account NAME]        list all mailboxes with unread/total counts
  messages
    list --account NAME --mailbox NAME [--limit N] [--unread] [--summary]
    read <message-id>
    open <message-id>
    search [--query TEXT] [--from ADDR] [--subject TEXT]
           [--account NAME] [--mailbox NAME] [--limit N] [--summary]
  send --to ADDR --subject SUBJ --body BODY
       [--to ADDR]... [--cc ADDR]... [--bcc ADDR]... [--attachment /abs/path]...
  rules
    list
    add-condition --rule NAME --expression VAL [--type TYPE] [--qualifier QUAL]
    create --name NAME --move MAILBOX --expression VAL [--account NAME]
           [--type TYPE] [--qualifier QUAL]
    apply [--rule NAME] [--mailbox NAME] [--account NAME] [--batch-size N]
  schema                            JSON spec of all commands and flags
```

**rules apply** has two modes:
- `--rule NAME` → targeted mode: uses indexed `whose` AppleScript clauses per condition, very fast on large mailboxes. Iterates all accounts if `--account` omitted.
- No `--rule` → general mode: batched `perform mail action with messages` calls from Go (not AppleScript loops), avoiding the 5-minute osascript timeout. Default `--batch-size 500`.

**Rule type mappings** (short → AppleScript enum):
```
from          → from header
to            → to header
cc            → cc header
subject       → subject header
any-recipient → any recipient
body          → message content
account       → account
```

**Qualifier mappings**:
```
contains     → does contain value
not-contains → does not contain value
begins-with  → begins with value
ends-with    → ends with value
equals       → equal to value
```

**Rule scale constraint:** Mail.app GUI degrades beyond ~80 conditions per rule. When a rule hits 80, create a numbered sibling (`services2`, `services3`, etc.).

**Message IDs** are opaque strings from Mail.app (`message id` property). They are never constructed manually — always obtained from `messages list` or `messages search` output.

**Search behavior:** Without `--mailbox`, the search script automatically skips Trash, Junk, Sent, Drafts, and All Mail. `--from` and `--subject` use Mail.app's indexed `whose` clause (fast). `--query` searches subject+sender+body (slow; always add `--mailbox` to scope).

## aical Command Structure

```
aical [--format json] [--pretty]
  calendars                         list all calendars (id, name, writable, description)
  events
    list  [--calendar NAME]
          [--today | --days N | --from YYYY-MM-DD [--to YYYY-MM-DD]]
          [--limit N]
    get   <uid> [--calendar NAME]
    create --calendar NAME --summary TITLE --start DATETIME --end DATETIME
           [--allday] [--location LOC] [--notes TEXT] [--url URL]
    update <uid> [--calendar NAME]
           [--summary TITLE] [--start DATETIME] [--end DATETIME]
           [--allday true|false] [--location LOC] [--notes TEXT] [--url URL]
    delete <uid> [--calendar NAME] --confirm
    open   <uid> [--calendar NAME]
```

**Date range defaults (list):** no flags → today through today+6 days. `--days N` → today through today+(N-1). `--today` → today only. `--from`/`--to` → explicit range.

**Event fields:** `uid`, `summary`, `start` (ISO 8601), `end`, `allDay`, `location`, `notes`, `status`, `url`, `calendar`. `get` also returns `attendees[]` (name, email, status).

**Attendee fields (read-only):** `name`, `email`, `status` (unknown/accepted/declined/tentative). Participation status cannot be changed via AppleScript — use `events open` to open in Calendar.app GUI for RSVP.

**`events open`:** Uses the AppleScript `show` command to navigate Calendar.app to the event and calls `activate`. For invite acceptance/declination by the user.

**`events delete`:** Requires `--confirm` flag to prevent accidental deletion.

**Event status field:** The `status` property on events is readable (none/confirmed/cancelled/tentative) but cannot be set on iCloud/CalDAV events via AppleScript (Calendar.app rejects it with error -10000). `--status` flag is not exposed.

**Calendar.app date construction in AppleScript:** Uses a `makeDateTime(y, m, d, secs)` handler that sets day to 1 before changing month (month-boundary overflow workaround), then sets `time of theDate to secs` (seconds since midnight). This is different from the `makeDate(y, m, d)` handler in `events_list.applescript` which only takes a date.

**`calendarIdentifier` property:** Unreliable on CalDAV/iCloud calendars (returns AppleEvent handler failed -10000). `calendars_list.applescript` falls back to `name` as the `id` value when `calendarIdentifier` fails.

**`asLiteral` template function:** Added to `internal/ical/client.go` `RenderScript`. Converts a Go string to a valid AppleScript string expression, escaping embedded `"` as `& (ASCII character 34) &`. Used in templates as `{{asLiteral .FieldName}}` for all user-supplied string inputs to the create/update/get/delete/open scripts.

## Codesigning

Both binaries require ad-hoc codesigning with the `com.apple.security.automation.apple-events` entitlement to receive TCC (Transparency, Consent, and Control) automation permission from macOS. Without this, osascript calls fail silently or prompt the user without the right app association.

The `entitlements.plist` at the repo root is used for both local builds and plugin binaries. The `plugins/*/bin/entitlements.plist` files are copies.

**Signing commands:**
```sh
codesign --sign - --force --entitlements entitlements.plist <binary>
```

`--sign -` = ad-hoc (no Developer ID required). This must be re-run after every copy of the binary to a new path, because codesigning is path-bound.

On first `osascript` invocation, macOS shows a one-time automation permission prompt. The user must click **Allow**.

## Build Targets

```makefile
make              # build + sign + install both to GOPATH/bin (default)
make build        # compile ./amail and ./aical (current platform)
make sign         # ad-hoc codesign local ./amail and ./aical
make install      # build + sign + copy to GOPATH/bin + re-sign in GOPATH/bin
make skill        # cross-compile arm64 plugin binaries to plugins/*/bin/ + sign
make skill-mail   # apple-mail plugin only
make skill-ical   # ical plugin only
make clean        # remove local ./amail ./aical and plugins/*/bin/ artifacts
make fmt          # go fmt ./...
make vet          # go vet ./...
```

`make skill` cross-compiles with `GOOS=darwin GOARCH=arm64` — these are the distribution binaries committed to `plugins/*/bin/`. `make install` compiles for the current host (also arm64 in practice) and installs to GOPATH/bin for local development use.

The plugin binaries in `plugins/*/bin/` are tracked in git. `.gitignore` only excludes the local build artifacts (`/amail`, `/aical` at repo root).

## Plugin Marketplace

This repo IS a Claude Code plugin marketplace. The catalog is at `.claude-plugin/marketplace.json`:

```json
{
  "name": "amail-plugins",
  "plugins": [
    { "name": "apple-mail", "source": "./plugins/apple-mail", ... },
    { "name": "ical",        "source": "./plugins/ical",        ... }
  ]
}
```

Each plugin directory has:
- `.claude-plugin/plugin.json` — plugin metadata (name, version, description, author)
- `bin/<binary>` — pre-built arm64 binary + `entitlements.plist`
- `skills/<name>/SKILL.md` — Claude Code skill definition

The plugin system is expected to add `plugins/*/bin/` to PATH when a plugin is installed.

**Install commands (user-facing):**
```sh
/plugin marketplace add trodemaster/apple-mail-cli
/plugin install apple-mail@amail-plugins
/plugin install ical@amail-plugins
```

## Skill Setup (Current State)

The plugin skills are mirrored to two locations that Claude Code reads from:
- `~/.claude/skills/apple-mail/SKILL.md` — global user skill
- `~/.claude/skills/ical/SKILL.md` — global user skill
- `~/Developer/machine-cfg/claude/skills/apple-mail/SKILL.md` — machine-cfg mirror
- `~/Developer/machine-cfg/claude/skills/ical/SKILL.md` — machine-cfg mirror

**Source of truth:** `plugins/apple-mail/skills/apple-mail/SKILL.md` and `plugins/ical/skills/ical/SKILL.md`. When the SKILL.md files change, copy them to all four locations.

The `amail-plugins` marketplace is registered in `~/.claude/settings.json` under `extraKnownMarketplaces` and in `~/.claude/plugins/known_marketplaces.json`. Both plugins are listed in `enabledPlugins` and `~/.claude/plugins/installed_plugins.json` with `installPath` pointing to the local plugin directories.

**Binaries:** Both `amail` and `aical` are installed to GOPATH/bin (`~/Developer/go/bin/`) via `make install`. GOPATH/bin is in PATH.

## Platform Constraints

- macOS 13+ (Ventura or later) required
- Apple Silicon (arm64) only — the plugin binaries are arm64; no x86_64 support
- Mail.app and Calendar.app must be running when the CLIs are invoked
- TCC automation permission must be granted once per binary path (re-grant required if binary moves)

## Extending the Codebase

**Adding a new amail command:**
1. Add a `.applescript` template to `internal/mail/scripts/`
2. Add a Go function in the appropriate `internal/mail/*.go` file using `//go:embed` and `RenderScript`/`RenderScriptResilient`
3. Add a cobra command in `cmd/` following the existing pattern (check `isJSON`, use `output.PrintJSON` or `output.PrintError`)
4. Register it in the appropriate `init()` with `rootCmd.AddCommand`
5. Update `plugins/apple-mail/skills/apple-mail/SKILL.md` to document the new command
6. Mirror the updated SKILL.md to `~/.claude/skills/apple-mail/SKILL.md` and `machine-cfg`
7. Run `make skill-mail` to rebuild the plugin binary, then commit the updated `plugins/apple-mail/bin/amail`

**Adding a new aical command:** Same pattern under `internal/ical/`, `ical/cmd/`, and `plugins/ical/`.

**Never use** `RunScript` for operations on large mailboxes — use `RenderScriptResilient` so Mail.app hangs are handled automatically. For rules apply on large mailboxes, batch from Go rather than looping in AppleScript.

**Template injection safety:** AppleScript templates receive user-supplied strings (account names, mailbox names, expressions). The scripts use AppleScript string literals — ensure any user input passed into a template is not used in a context where AppleScript interprets it as code. The current pattern embeds values as AppleScript string literals (`set x to "{{.Value}}"`); the `escapeJSON` handler in each script only handles JSON output, not input escaping. If adding new parameters, verify they cannot break out of AppleScript string context.
