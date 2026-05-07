---
name: ical
description: "Read and manage Apple Calendar.app events on macOS Apple Silicon. Use when the user asks about their schedule, wants to see upcoming events, create a new event, update or delete an existing event, accept/decline invites, check what's on today, find events in a date range, or list their calendars. Requires: macOS on Apple Silicon (arm64)."
argument-hint: "calendar task — e.g. 'what's on today', 'create meeting Friday 2pm', 'open invite to accept it'"
allowed-tools: Bash(aical *)
---

# ical

Full management of Apple Calendar.app via the bundled `aical` CLI (Apple Silicon arm64, no credentials, no CalDAV API keys). The plugin system adds `aical` to PATH automatically.

## Requirements

- macOS 13+ on Apple Silicon (arm64)
- Calendar.app open and running
- Plugin installed via the Claude Code marketplace

On first `aical` invocation, macOS will prompt for Calendar.app automation permission — tell the user to click **Allow**.

## List calendars

```
aical calendars
```

Returns: `id`, `name`, `writable`, `description` per calendar. Run this first if you don't know a calendar name — never guess.

## List events

```
aical events list [--calendar "Name"] [--today] [--days N] \
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

## Get full event details (with attendees)

```
aical events get <uid> [--calendar "Name"]
```

Returns complete event detail including the `attendees` array (name, email, participation status).
`--calendar` is optional but speeds up the search when provided.

Use this to inspect an invite: check who is attending and their RSVP status.

## Create an event

```
aical events create --calendar "Name" --summary "Title" \
  --start YYYY-MM-DDTHH:MM:SS --end YYYY-MM-DDTHH:MM:SS \
  [--allday] [--location "..."] [--notes "..."] [--url "..."]
```

- `--calendar` must be a **writable** calendar (check `aical calendars` for `writable: true`)
- Date format: `2026-05-08T14:00:00` for timed events, `2026-05-08` for all-day
- `--allday` marks the event as all-day (time components ignored)
- Returns: `uid`, `summary`, `start`, `end`, `allDay`, `calendar`

**Before creating:** show the user a plain-English summary (calendar, title, date/time) and confirm.

## Update an event

```
aical events update <uid> [--calendar "Name"] \
  [--summary "New Title"] \
  [--start YYYY-MM-DDTHH:MM:SS] [--end YYYY-MM-DDTHH:MM:SS] \
  [--allday true|false] \
  [--location "..."] [--notes "..."] [--url "..."] \
  [--status confirmed|cancelled|tentative|none]
```

Only the flags you provide are changed. `--calendar` restricts the search to one calendar (faster).

**Before updating:** show the current event details and the proposed changes, then confirm.

## Delete an event

```
aical events delete <uid> [--calendar "Name"] --confirm
```

`--confirm` is required to prevent accidental deletion. **Show the event summary to the user and ask before calling this.**

## Open an event in Calendar.app (accept/decline invites)

```
aical events open <uid> [--calendar "Name"]
```

Opens the event in the Calendar.app GUI and brings Calendar.app to the front. Use this when the user wants to accept or decline an invite — Calendar.app shows the RSVP buttons. This is the correct way to handle invite responses (participation status cannot be changed via AppleScript).

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
aical events list --today --format json
aical events list --from 2026-05-01 --to 2026-05-31 --format json | jq '.data[] | select(.calendar == "Work") | .summary'
aical events get <uid> --format json | jq '.data.attendees'
aical calendars --format json | jq '.data[].name'
```

## Common patterns

**What's on today?**
```
aical events list --today
```

**This week:**
```
aical events list --days 7
```

**Specific calendar, next 30 days:**
```
aical events list --calendar "Work" --days 30
```

**Who's invited to a meeting?**
```
aical events get <uid> --format json | jq '.data.attendees[] | {name, email, status}'
```

**Create a 1-hour meeting:**
```
aical events create --calendar "Work" --summary "1:1 with Alice" \
  --start 2026-05-09T14:00:00 --end 2026-05-09T15:00:00
```

**Reschedule a meeting:**
```
aical events update <uid> --start 2026-05-09T15:00:00 --end 2026-05-09T16:00:00
```

**Accept/decline an invite:**
```
aical events open <uid>
```
Then use the RSVP buttons in Calendar.app.

## Guidelines

- Always run `aical calendars` first if you don't know a calendar name — never guess it
- `--calendar` on get/update/delete/open speeds up the UID search; provide it when you know the calendar
- Use `--format json` whenever you need to extract specific fields or filter output
- Before creating or updating, show the user a plain-English summary and confirm
- `delete` is irreversible — always confirm with the user before calling with `--confirm`
- Participation status (accept/decline) is read-only via AppleScript; use `events open` for invite handling
- If `ok` is `false`, surface the `error.message` to the user and stop
- Calendar.app must be open; if an osascript error occurs, ask the user to open Calendar.app
