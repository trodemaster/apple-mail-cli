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

// asLiteral converts a Go string into a valid AppleScript string expression,
// escaping embedded double-quote characters using (ASCII character 34).
func asLiteral(s string) string {
	parts := strings.Split(s, `"`)
	pieces := make([]string, len(parts))
	for i, p := range parts {
		pieces[i] = `"` + p + `"`
	}
	return strings.Join(pieces, ` & (ASCII character 34) & `)
}

// RenderScript renders a text/template with data and runs it via osascript.
func RenderScript(scriptTmpl string, data interface{}) (string, error) {
	funcMap := template.FuncMap{"asLiteral": asLiteral}
	tmpl, err := template.New("script").Funcs(funcMap).Parse(scriptTmpl)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}
	return RunScript(buf.String())
}
