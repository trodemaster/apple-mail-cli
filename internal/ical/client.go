package ical

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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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

// RenderScript renders a text/template with data and runs it via osascript.
func RenderScript(scriptTmpl string, data interface{}) (string, error) {
	tmpl, err := template.New("script").Parse(scriptTmpl)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return RunScript(buf.String())
}
