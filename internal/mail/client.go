package mail

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

// RunScript executes an AppleScript string via osascript and returns stdout.
func RunScript(script string) (string, error) {
	return runScriptTimeout(script, 5*time.Minute)
}

func runScriptTimeout(script string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("osascript: %s", msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// isMailHung returns true if the error looks like Mail.app is unresponsive.
func isMailHung(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "AppleEvent timed out") ||
		strings.Contains(msg, "signal: killed") ||
		strings.Contains(msg, "Connection is invalid")
}

// RestartMail kills Mail.app and relaunches it, waiting up to 15 seconds for
// it to become responsive before returning.
func RestartMail() error {
	// Kill any running Mail.app process.
	_ = exec.Command("pkill", "-x", "Mail").Run()
	time.Sleep(2 * time.Second)

	// Relaunch.
	if err := exec.Command("open", "-a", "Mail").Run(); err != nil {
		return fmt.Errorf("relaunch Mail: %w", err)
	}

	// Wait for Mail.app to respond to Apple Events (up to 15 seconds).
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		_, err := runScriptTimeout(`tell application "Mail" to get name of first account`, 5*time.Second)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("Mail.app did not become responsive after restart")
}

// RunScriptResilient runs a script with a 120-second timeout per attempt.
// If Mail.app hangs, it kills and relaunches it, then retries once.
// The onRestart callback (if non-nil) is called before the restart attempt.
func RunScriptResilient(script string, onRestart func()) (string, error) {
	const attemptTimeout = 120 * time.Second

	out, err := runScriptTimeout(script, attemptTimeout)
	if err == nil {
		return out, nil
	}

	if !isMailHung(err) {
		return "", err
	}

	// Mail.app appears hung — restart and retry once.
	if onRestart != nil {
		onRestart()
	}
	if restartErr := RestartMail(); restartErr != nil {
		return "", fmt.Errorf("mail hang detected; restart failed: %w", restartErr)
	}

	// One retry after restart.
	return runScriptTimeout(script, attemptTimeout)
}

// RenderScript renders a text/template with data and runs it via osascript.
func RenderScript(scriptTmpl string, data interface{}) (string, error) {
	return RenderScriptResilient(scriptTmpl, data, nil)
}

// RenderScriptResilient renders and runs a script with hang detection and
// automatic Mail.app restart on timeout. The onRestart callback is optional.
func RenderScriptResilient(scriptTmpl string, data interface{}, onRestart func()) (string, error) {
	tmpl, err := template.New("script").Parse(scriptTmpl)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return RunScriptResilient(buf.String(), onRestart)
}
