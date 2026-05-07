package ical

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"
)

//go:embed scripts/events_list.applescript
var eventsListScript string

//go:embed scripts/events_get.applescript
var eventsGetScript string

//go:embed scripts/events_create.applescript
var eventsCreateScript string

//go:embed scripts/events_update.applescript
var eventsUpdateScript string

//go:embed scripts/events_delete.applescript
var eventsDeleteScript string

//go:embed scripts/events_open.applescript
var eventsOpenScript string

// Attendee represents a calendar event attendee.
type Attendee struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

// Event represents a Calendar.app event.
type Event struct {
	UID       string     `json:"uid"`
	Summary   string     `json:"summary"`
	Start     string     `json:"start"`
	End       string     `json:"end"`
	AllDay    bool       `json:"allDay"`
	Location  string     `json:"location,omitempty"`
	Notes     string     `json:"notes,omitempty"`
	Status    string     `json:"status,omitempty"`
	URL       string     `json:"url,omitempty"`
	Calendar  string     `json:"calendar"`
	Attendees []Attendee `json:"attendees,omitempty"`
}

// ListEventsParams configures an event query.
type ListEventsParams struct {
	CalendarName string
	From         time.Time
	To           time.Time
	Limit        int
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

type getEventParams struct {
	UID          string
	CalendarName string
}

// GetEvent retrieves full details for a single event by UID, including attendees.
// CalendarName is optional; if set it restricts the search to one calendar (faster).
func GetEvent(uid, calendarName string) (*Event, error) {
	out, err := RenderScript(eventsGetScript, getEventParams{UID: uid, CalendarName: calendarName})
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	var evt Event
	if err := json.Unmarshal([]byte(out), &evt); err != nil {
		return nil, fmt.Errorf("parse event: %w (raw: %s)", err, out)
	}
	return &evt, nil
}

// CreateEventParams defines a new event to create.
type CreateEventParams struct {
	CalendarName string
	Summary      string
	StartYear, StartMonth, StartDay int
	StartSecs    int
	EndYear, EndMonth, EndDay int
	EndSecs      int
	AllDay       bool
	Location     string
	Notes        string
	URL          string
}

// ParseEventTime parses an ISO 8601 datetime string (YYYY-MM-DDTHH:MM:SS or YYYY-MM-DD)
// into year/month/day and seconds-since-midnight components.
func ParseEventTime(s string) (year, month, day, secs int, err error) {
	var t time.Time
	if len(s) == 10 {
		t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	} else {
		t, err = time.ParseInLocation("2006-01-02T15:04:05", s, time.Local)
	}
	if err != nil {
		return 0, 0, 0, 0, err
	}
	h, m, sc := t.Clock()
	return t.Year(), int(t.Month()), t.Day(), h*3600 + m*60 + sc, nil
}

// CreateEvent creates a new event in Calendar.app.
func CreateEvent(p CreateEventParams) (*Event, error) {
	out, err := RenderScript(eventsCreateScript, p)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	var evt Event
	if err := json.Unmarshal([]byte(out), &evt); err != nil {
		return nil, fmt.Errorf("parse created event: %w (raw: %s)", err, out)
	}
	return &evt, nil
}

// UpdateEventParams specifies which fields to change on an existing event.
// Set the corresponding SetXxx bool to true to apply that field's value.
type UpdateEventParams struct {
	UID          string
	CalendarName string
	Summary      string
	SetSummary   bool
	Location     string
	SetLocation  bool
	Notes        string
	SetNotes     bool
	URL          string
	SetURL       bool
	Status       string // confirmed | cancelled | tentative | none
	SetStatus    bool
	AllDay       bool
	SetAllDay    bool
	StartYear, StartMonth, StartDay int
	StartSecs    int
	SetStart     bool
	EndYear, EndMonth, EndDay int
	EndSecs      int
	SetEnd       bool
}

// UpdateEvent modifies an existing event identified by UID.
func UpdateEvent(p UpdateEventParams) error {
	_, err := RenderScript(eventsUpdateScript, p)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

type uidCalParams struct {
	UID          string
	CalendarName string
}

// DeleteEvent removes an event from Calendar.app by UID.
// CalendarName is optional; if set it restricts the search to one calendar (faster).
func DeleteEvent(uid, calendarName string) error {
	_, err := RenderScript(eventsDeleteScript, uidCalParams{UID: uid, CalendarName: calendarName})
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return nil
}

// OpenEvent opens an event in the Calendar.app GUI and brings Calendar.app to the front.
// CalendarName is optional; if set it restricts the search to one calendar (faster).
func OpenEvent(uid, calendarName string) error {
	_, err := RenderScript(eventsOpenScript, uidCalParams{UID: uid, CalendarName: calendarName})
	if err != nil {
		return fmt.Errorf("open event: %w", err)
	}
	return nil
}
