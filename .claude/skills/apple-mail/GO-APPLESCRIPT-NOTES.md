# Go + AppleScript Compatibility Notes

Lessons learned from building `amail` — a Go CLI that drives Mail.app via `osascript` subprocess.

---

## 1. AppleScript does NOT support backslash escape sequences

**The most important lesson.** The `\"` escape does NOT work inside AppleScript string literals when the script is executed via `osascript -e` or from an embedded file. AppleScript has no backslash escaping — `\` is a literal character, not an escape prefix.

**Wrong (will cause syntax error):**
```applescript
if c = "\"" then          -- AppleScript reads this as string "\" then stray "
    set s to s & "\\\""   -- broken
end if
```

**Correct — use ASCII character constants:**
```applescript
set q  to (ASCII character 34)   -- double quote "
set bs to (ASCII character 92)   -- backslash \

if c = q then
    set s to s & bs & q          -- outputs \"
end if
```

You can also use the built-in `quote` constant for `"`, but it does not help with backslash. For JSON generation, always use the `q`/`bs` variable pattern above.

> **Note:** The AppleScript skill's operator table lists `\"` as a valid escape. This is **incorrect for osascript execution** — it only works inside Script Editor's compiled `.scpt` format in some versions. Never rely on it when running scripts through `osascript`.

---

## 2. `first` is a reserved keyword — never use it as a variable name

AppleScript reserves `first` as a positional reference (e.g. `first item of list`). Using it as a loop sentinel causes a cryptic parse error:

```
syntax error: Expected class name but found "to". (-2741)
```

The error triggers at `set first to true` because the parser sees `first` as a class reference and then hits `to` where it expects something else.

**Wrong:**
```applescript
set first to true
if not first then ...
```

**Correct:**
```applescript
set isFirst to true
if not isFirst then ...
```

Other risky single-word variable names to avoid: `last`, `front`, `back`, `result`, `it`, `me`, `my`, `its`, `count`, `id`, `name`, `class`, `index`.

---

## 3. `//go:embed` cannot use `../` relative paths

Go's embed directive requires the embedded path to be within or below the package directory. You cannot reference files in a parent directory.

**Wrong (package in `internal/mail/`, scripts at project root `scripts/`):**
```go
//go:embed ../../scripts/accounts_list.applescript  // compile error
```

**Correct — scripts must live inside the package tree:**
```
internal/mail/scripts/accounts_list.applescript
```
```go
//go:embed scripts/accounts_list.applescript
var accountsListScript string
```

---

## 4. Call local handlers with `my` inside `tell application` blocks

Inside a `tell application "Mail"` block, bare function calls are dispatched to Mail.app, not your script. Prefix local handler calls with `my`:

```applescript
tell application "Mail"
    -- WRONG: AppleScript tries to send escapeJSON to Mail.app
    set s to escapeJSON(name of acc)

    -- CORRECT: calls the handler defined in this script
    set s to my escapeJSON(name of acc)
end tell
```

---

## 5. Go `text/template` and AppleScript work well together, with caveats

The pattern of embedding `.applescript` files and rendering them with `text/template` is clean. Key rules:

- Template delimiters `{{` and `}}` are safe as long as your AppleScript doesn't contain those literal sequences (uncommon in practice)
- AppleScript booleans must be lowercase `true`/`false` — Go's `fmt.Sprintf("%v", bool)` produces the right output; template rendering of a `bool` field does too
- String parameters injected via template (e.g. `"{{.AccountName}}"`) are **not** automatically escaped — if a user-controlled string contains `"`, it will break the rendered script. For now inputs are trusted; a future hardening step would be to sanitize or use a different injection mechanism

```go
// Boolean fields render correctly via template
type params struct {
    UnreadOnly string  // pass "true" or "false" as string, not bool
}
```

---

## 6. `osascript -e` takes the full script as a single argument

Go's `os/exec` passes the script string directly — no shell quoting involved. This means:
- No shell injection risk from the script content itself
- The script is limited by `ARG_MAX` (~2MB on macOS) — fine for embedded scripts, not for dynamically constructed giant scripts
- Newlines in the script string are preserved and work correctly

```go
cmd := exec.CommandContext(ctx, "osascript", "-e", scriptString)
```

---

## 7. AppleScript boolean-to-string coercion produces valid JSON booleans

```applescript
set isRead to read status of msg   -- boolean
(isRead as string)                 -- produces "true" or "false"
```

Go's `json.Unmarshal` into a `bool` field handles `true`/`false` JSON values correctly. No conversion needed.

---

## 8. `count` collisions

`count` is both a command (`count of list`) and a potential variable name conflict. Safe usage:

```applescript
-- OK: count as command with explicit operand
set n to count of (messages of mbox)

-- Risky: using `count` as a variable name shadows the command
set count to 0   -- avoid; use msgCount, totalCount, etc.
```

In `messages_list.applescript` we originally used `count` as a loop variable — renamed to `msgCount` to avoid any ambiguity.

---

## 9. `(boolVar as string)` vs JSON `true`/`false` casing

AppleScript's `as string` on booleans yields lowercase `"true"`/`"false"`, which matches JSON exactly. This is a happy coincidence — no post-processing needed when assembling JSON strings from AppleScript.

---

## 11. Mail.app rule qualifier enum constants conflict with AppleScript operators

When setting a `rule condition` qualifier property in a record literal, multi-word constants that overlap with AppleScript operators are silently misinterpreted:

```applescript
-- WRONG: "ends with" is an AppleScript string operator; Mail.app ignores it and defaults to "does contain value"
make new rule condition ... with properties {rule type:from header, qualifier:ends with value, expression:"domain.com"}
```

**Correct — set the qualifier as a separate statement AFTER creation:**
```applescript
set newCond to make new rule condition at end of rule conditions of theRule ¬
    with properties {rule type:from header, expression:"domain.com"}
set qualifier of newCond to ends with value   -- works correctly here
```

This applies to any qualifier that contains `ends with`, `begins with`, or other operator-like words. The `rule type` property does not have this problem because `from header`, `subject header`, etc. don't overlap with AppleScript operators.

---

## 13. Pass `whose` specifiers directly to `move` — never store them in a variable first

Storing a `whose`-filtered result in a variable materializes it as a brace list `{msg1, msg2, ...}`. Mail.app's `move` command does not accept a brace list — it expects either a single message specifier or an object specifier expression.

**Wrong:**
```applescript
set batch to (messages of mbox whose sender ends with "domain.com>")
move batch to destMailbox   -- error: Can't make {message id...} into type specifier (-1700)
```

**Correct — pass the specifier directly:**
```applescript
set n to count of (messages of mbox whose sender ends with "domain.com>")
if n > 0 then
    move (messages of mbox whose sender ends with "domain.com>") to destMailbox
end if
```

The double evaluation (count + move) is necessary because `move` with zero results throws an error.

---

## 14. Mail.app `sender` property includes display name and angle brackets

The `sender` property of a message is the raw From header value: `"Display Name <email@domain.com>"`. This affects `whose sender` comparisons:

- `whose sender ends with "domain.com"` → **0 results** (string ends with `>`, not the domain)
- `whose sender ends with "domain.com>"` → **works** — matches the angle-bracket format
- `whose sender contains "domain.com"` → **works** — matches anywhere in the string
- `whose sender contains "@domain.com"` → **more precise** — matches only the email address part

Mail.app's rule engine strips the display name before applying `from ends-with` — but AppleScript's `whose` does not. When replicating rule behavior in `whose` clauses, append `">"` for `ends with` comparisons on from/sender fields.

---

## 12. `delete rule condition` invalidates Mail.app's Apple Events connection

Calling `delete` on a `rule condition` object causes Mail.app to return error `-609` (Connection is invalid) on the same or subsequent Apple Events calls within the same script session. This is a Mail.app bug.

**Workaround:** Avoid deleting individual conditions via AppleScript. Instead:
- Correct a condition in-place by using `set qualifier of cond to ...` / `set expression of cond to ...`
- If the rule must be rebuilt, do it in the Mail.app GUI (Preferences → Rules)
- Appending new conditions with `make new rule condition` works fine; only `delete` is problematic

---

## 10. `osascript` exit code on AppleScript errors

When a script fails at runtime (e.g. mailbox not found), `osascript` exits with code 1 and writes the error to stderr. The Go client should capture stderr and surface it:

```go
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
if err := cmd.Run(); err != nil {
    msg := strings.TrimSpace(stderr.String())
    return "", fmt.Errorf("osascript: %s", msg)
}
```

Syntax errors in the script also produce exit code 1 with a character-offset error in stderr like `494:496: syntax error: ...` — useful for pinpointing the broken line.
