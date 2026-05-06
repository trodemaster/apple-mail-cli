package ical

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed scripts/calendars_list.applescript
var calendarsListScript string

// Calendar represents a Calendar.app calendar.
type Calendar struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Writable    bool   `json:"writable"`
	Description string `json:"description"`
}

// ListCalendars returns all calendars configured in Calendar.app.
func ListCalendars() ([]Calendar, error) {
	out, err := RunScript(calendarsListScript)
	if err != nil {
		return nil, fmt.Errorf("list calendars: %w", err)
	}
	var cals []Calendar
	if err := json.Unmarshal([]byte(out), &cals); err != nil {
		return nil, fmt.Errorf("parse calendars: %w (raw: %s)", err, out)
	}
	return cals, nil
}
