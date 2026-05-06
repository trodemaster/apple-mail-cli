package ical

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed scripts/events_list.applescript
var eventsListScript string

// Event represents a Calendar.app event.
type Event struct {
	UID      string `json:"uid"`
	Summary  string `json:"summary"`
	Start    string `json:"start"`
	End      string `json:"end"`
	AllDay   bool   `json:"allDay"`
	Location string `json:"location,omitempty"`
	Notes    string `json:"notes,omitempty"`
	Status   string `json:"status,omitempty"`
	URL      string `json:"url,omitempty"`
	Calendar string `json:"calendar"`
}

// ListEventsParams configures an event query.
type ListEventsParams struct {
	// CalendarName filters to a single calendar; empty means all calendars.
	CalendarName string
	From         time.Time
	To           time.Time
	// Limit caps results; 0 means no limit.
	Limit int
}

type eventsScriptParams struct {
	CalendarName string
	FromYear     int
	FromMonth    int
	FromDay      int
	ToYear       int
	ToMonth      int
	ToDay        int
	Limit        int
}

// ListEvents queries Calendar.app for events within the given date range.
func ListEvents(p ListEventsParams) ([]Event, error) {
	out, err := RenderScript(eventsListScript, eventsScriptParams{
		CalendarName: p.CalendarName,
		FromYear:     p.From.Year(),
		FromMonth:    int(p.From.Month()),
		FromDay:      p.From.Day(),
		ToYear:       p.To.Year(),
		ToMonth:      int(p.To.Month()),
		ToDay:        p.To.Day(),
		Limit:        p.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	var events []Event
	if err := json.Unmarshal([]byte(out), &events); err != nil {
		return nil, fmt.Errorf("parse events: %w (raw: %s)", err, out)
	}
	return events, nil
}
