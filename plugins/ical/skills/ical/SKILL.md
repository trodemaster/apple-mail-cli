---
name: ical
description: "Read Apple Calendar events and calendars on macOS Apple Silicon. Use when the user asks about their schedule, wants to see upcoming events, check what's on today, find events in a date range, list their calendars, or anything related to Calendar.app or iCal. Read-only. Requires: macOS on Apple Silicon (arm64)."
argument-hint: "calendar task — e.g. 'what's on today', 'events this week', 'list calendars'"
allowed-tools: Bash(aical *)
---

# ical

Read-only access to Apple Calendar.app via the bundled `aical` CLI (Apple Silicon arm64, no credentials, no CalDAV API keys). The plugin system adds `aical` to PATH automatically.

## Requirements

- macOS 13+ on Apple Silicon (arm64)
- Calendar.app open and running
- Plugin installed via the Claude Code marketplace

On first `aical` invocation, macOS will prompt for Calendar.app automation permission — tell the user to click **Allow**.

## List calendars

```
aical calendars
```

Returns: `id`, `name`, `writable`, `description` per calendar.

## List events

```
aical events [--calendar "Name"] [--today] [--days N] \
  [--from YYYY-MM-DD] [--to YYYY-MM-DD] \
  [--limit N]
```

- Default: events starting today through the next 7 days, all calendars
- `--today` — only today's events
- `--days N` — next N days from today (default 7)
- `--from` / `--to` — explicit date range (YYYY-MM-DD)
- `--calendar` — restrict to one calendar by exact name
- `--limit` — max events returned (default 50; 0 = unlimited)

Output fields per event: `uid`, `summary`, `start`, `end`, `allDay`, `location`, `notes`, `status`, `url`, `calendar`

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

Use `--pretty` for formatted JSON.

```
aical events --today --format json
aical events --from 2026-05-01 --to 2026-05-31 --format json | jq '.data[] | select(.calendar == "Work") | .summary'
aical calendars --format json | jq '.data[].name'
```

## Common patterns

**What's on today?**
```
aical events --today
```

**This week:**
```
aical events --days 7
```

**Specific calendar, next 30 days:**
```
aical events --calendar "Work" --days 30
```

**Date range:**
```
aical events --from 2026-06-01 --to 2026-06-30 --calendar "Personal"
```

## Guidelines

- Always run `aical calendars` first if you don't know a calendar name — never guess it
- For "what's on today" / "do I have any meetings" — use `--today`
- For "what's this week" — use `--days 7` (default)
- `--format json` whenever you need to extract specific fields or filter
- If `ok` is `false`, surface the `error.message` to the user and stop
- Calendar.app must be open; if an osascript error occurs, ask the user to open Calendar.app
