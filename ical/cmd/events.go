package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/trodemaster/apple-mail-cli/internal/ical"
	"github.com/trodemaster/apple-mail-cli/internal/output"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage calendar events: list, get, create, update, delete, open",
}

// --- list ---

var (
	eventsListCalendar string
	eventsListFrom     string
	eventsListTo       string
	eventsListToday    bool
	eventsListDays     int
	eventsListLimit    int
)

var eventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List events in a date range",
	Long: `List events from Calendar.app within a date range.

Defaults to events starting today through the next 7 days across all calendars.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		var from, to time.Time
		switch {
		case eventsListToday:
			from = today
			to = today
		case eventsListFrom != "" || eventsListTo != "":
			var err error
			if eventsListFrom != "" {
				from, err = time.ParseInLocation("2006-01-02", eventsListFrom, now.Location())
				if err != nil {
					output.PrintError("events list", "invalidDate", fmt.Sprintf("--from: %s", err), prettyFlag)
					return nil
				}
			} else {
				from = today
			}
			if eventsListTo != "" {
				to, err = time.ParseInLocation("2006-01-02", eventsListTo, now.Location())
				if err != nil {
					output.PrintError("events list", "invalidDate", fmt.Sprintf("--to: %s", err), prettyFlag)
					return nil
				}
			} else {
				to = from.AddDate(0, 0, 6)
			}
		default:
			from = today
			to = today.AddDate(0, 0, eventsListDays-1)
		}

		events, err := ical.ListEvents(ical.ListEventsParams{
			CalendarName: eventsListCalendar,
			From:         from,
			To:           to,
			Limit:        eventsListLimit,
		})
		if err != nil {
			output.PrintError("events list", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events list", events, prettyFlag)
			return nil
		}

		t := output.NewTable("DATE", "TIME", "CALENDAR", "SUMMARY", "LOCATION")
		for _, e := range events {
			dateStr, timeStr := formatEventTime(e.Start, e.AllDay)
			t.AddRow(dateStr, timeStr, e.Calendar, e.Summary, e.Location)
		}
		fmt.Printf("%d event(s)  [%s → %s]\n\n",
			len(events),
			from.Format("Jan 2"),
			to.Format("Jan 2, 2006"))
		t.Print()
		return nil
	},
}

// --- get ---

var (
	eventsGetCalendar string
)

var eventsGetCmd = &cobra.Command{
	Use:   "get <uid>",
	Short: "Get full details for an event by UID (includes attendees)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		evt, err := ical.GetEvent(args[0], eventsGetCalendar)
		if err != nil {
			output.PrintError("events get", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events get", evt, prettyFlag)
			return nil
		}

		dateStr, timeStr := formatEventTime(evt.Start, evt.AllDay)
		fmt.Printf("UID:      %s\n", evt.UID)
		fmt.Printf("Summary:  %s\n", evt.Summary)
		fmt.Printf("Date:     %s  %s\n", dateStr, timeStr)
		_, endTime := formatEventTime(evt.End, evt.AllDay)
		if endTime != "" && endTime != "all-day" {
			fmt.Printf("End:      %s\n", endTime)
		}
		fmt.Printf("Calendar: %s\n", evt.Calendar)
		if evt.Location != "" {
			fmt.Printf("Location: %s\n", evt.Location)
		}
		if evt.Status != "" {
			fmt.Printf("Status:   %s\n", evt.Status)
		}
		if evt.URL != "" {
			fmt.Printf("URL:      %s\n", evt.URL)
		}
		if len(evt.Attendees) > 0 {
			fmt.Printf("Attendees:\n")
			for _, a := range evt.Attendees {
				name := a.Name
				if name == "" {
					name = a.Email
				}
				fmt.Printf("  %-30s  %-26s  %s\n", name, a.Email, a.Status)
			}
		}
		if evt.Notes != "" {
			fmt.Printf("\nNotes:\n%s\n", evt.Notes)
		}
		return nil
	},
}

// --- create ---

var (
	eventsCreateCalendar string
	eventsCreateSummary  string
	eventsCreateStart    string
	eventsCreateEnd      string
	eventsCreateAllDay   bool
	eventsCreateLocation string
	eventsCreateNotes    string
	eventsCreateURL      string
)

var eventsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new calendar event",
	RunE: func(cmd *cobra.Command, args []string) error {
		if eventsCreateCalendar == "" || eventsCreateSummary == "" || eventsCreateStart == "" || eventsCreateEnd == "" {
			output.PrintError("events create", "missingFlags", "--calendar, --summary, --start, and --end are required", prettyFlag)
			return nil
		}

		sy, sm, sd, ss, err := ical.ParseEventTime(eventsCreateStart)
		if err != nil {
			output.PrintError("events create", "invalidDate", fmt.Sprintf("--start: %s", err), prettyFlag)
			return nil
		}
		ey, em, ed, es, err := ical.ParseEventTime(eventsCreateEnd)
		if err != nil {
			output.PrintError("events create", "invalidDate", fmt.Sprintf("--end: %s", err), prettyFlag)
			return nil
		}

		evt, err := ical.CreateEvent(ical.CreateEventParams{
			CalendarName: eventsCreateCalendar,
			Summary:      eventsCreateSummary,
			StartYear:    sy, StartMonth: sm, StartDay: sd, StartSecs: ss,
			EndYear:      ey, EndMonth: em, EndDay: ed, EndSecs: es,
			AllDay:       eventsCreateAllDay,
			Location:     eventsCreateLocation,
			Notes:        eventsCreateNotes,
			URL:          eventsCreateURL,
		})
		if err != nil {
			output.PrintError("events create", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events create", evt, prettyFlag)
			return nil
		}

		fmt.Printf("Created: %s\n", evt.UID)
		dateStr, timeStr := formatEventTime(evt.Start, evt.AllDay)
		fmt.Printf("  %s  %s  %s\n", dateStr, timeStr, evt.Summary)
		return nil
	},
}

// --- update ---

var (
	eventsUpdateCalendar string
	eventsUpdateSummary  string
	eventsUpdateStart    string
	eventsUpdateEnd      string
	eventsUpdateAllDay   string // "true", "false", or ""
	eventsUpdateLocation string
	eventsUpdateNotes    string
	eventsUpdateURL      string
)

var eventsUpdateCmd = &cobra.Command{
	Use:   "update <uid>",
	Short: "Update fields on an existing event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p := ical.UpdateEventParams{
			UID:          args[0],
			CalendarName: eventsUpdateCalendar,
		}

		if cmd.Flags().Changed("summary") {
			p.Summary = eventsUpdateSummary
			p.SetSummary = true
		}
		if cmd.Flags().Changed("location") {
			p.Location = eventsUpdateLocation
			p.SetLocation = true
		}
		if cmd.Flags().Changed("notes") {
			p.Notes = eventsUpdateNotes
			p.SetNotes = true
		}
		if cmd.Flags().Changed("url") {
			p.URL = eventsUpdateURL
			p.SetURL = true
		}
		if cmd.Flags().Changed("allday") {
			p.AllDay = eventsUpdateAllDay == "true"
			p.SetAllDay = true
		}
		if cmd.Flags().Changed("start") {
			sy, sm, sd, ss, err := ical.ParseEventTime(eventsUpdateStart)
			if err != nil {
				output.PrintError("events update", "invalidDate", fmt.Sprintf("--start: %s", err), prettyFlag)
				return nil
			}
			p.StartYear, p.StartMonth, p.StartDay, p.StartSecs = sy, sm, sd, ss
			p.SetStart = true
		}
		if cmd.Flags().Changed("end") {
			ey, em, ed, es, err := ical.ParseEventTime(eventsUpdateEnd)
			if err != nil {
				output.PrintError("events update", "invalidDate", fmt.Sprintf("--end: %s", err), prettyFlag)
				return nil
			}
			p.EndYear, p.EndMonth, p.EndDay, p.EndSecs = ey, em, ed, es
			p.SetEnd = true
		}

		if err := ical.UpdateEvent(p); err != nil {
			output.PrintError("events update", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events update", map[string]string{"uid": args[0], "status": "updated"}, prettyFlag)
			return nil
		}
		fmt.Printf("Updated: %s\n", args[0])
		return nil
	},
}

// --- delete ---

var (
	eventsDeleteCalendar string
	eventsDeleteConfirm  bool
)

var eventsDeleteCmd = &cobra.Command{
	Use:   "delete <uid>",
	Short: "Delete an event by UID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !eventsDeleteConfirm {
			output.PrintError("events delete", "confirmRequired", "pass --confirm to permanently delete this event", prettyFlag)
			return nil
		}

		if err := ical.DeleteEvent(args[0], eventsDeleteCalendar); err != nil {
			output.PrintError("events delete", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events delete", map[string]string{"uid": args[0], "status": "deleted"}, prettyFlag)
			return nil
		}
		fmt.Printf("Deleted: %s\n", args[0])
		return nil
	},
}

// --- open ---

var (
	eventsOpenCalendar string
)

var eventsOpenCmd = &cobra.Command{
	Use:   "open <uid>",
	Short: "Open an event in Calendar.app (for accepting/declining invites)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ical.OpenEvent(args[0], eventsOpenCalendar); err != nil {
			output.PrintError("events open", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events open", map[string]string{"uid": args[0], "status": "opened"}, prettyFlag)
			return nil
		}
		fmt.Println("Opened in Calendar.app")
		return nil
	},
}

// formatEventTime splits an ISO datetime string into date and time components.
func formatEventTime(iso string, allDay bool) (date, timeStr string) {
	t, err := time.ParseInLocation("2006-01-02T15:04:05", iso, time.Local)
	if err != nil {
		return iso, ""
	}
	date = t.Format("Mon Jan 2")
	if allDay {
		return date, "all-day"
	}
	return date, t.Format("3:04 PM")
}

func init() {
	// list
	eventsListCmd.Flags().StringVar(&eventsListCalendar, "calendar", "", "filter to a specific calendar name")
	eventsListCmd.Flags().StringVar(&eventsListFrom, "from", "", "start date YYYY-MM-DD (default: today)")
	eventsListCmd.Flags().StringVar(&eventsListTo, "to", "", "end date YYYY-MM-DD (default: from + 6 days)")
	eventsListCmd.Flags().BoolVar(&eventsListToday, "today", false, "show only today's events")
	eventsListCmd.Flags().IntVar(&eventsListDays, "days", 7, "number of days from today")
	eventsListCmd.Flags().IntVar(&eventsListLimit, "limit", 50, "max events to return (0 = unlimited)")

	// get
	eventsGetCmd.Flags().StringVar(&eventsGetCalendar, "calendar", "", "restrict search to this calendar (faster)")

	// create
	eventsCreateCmd.Flags().StringVar(&eventsCreateCalendar, "calendar", "", "calendar name (required)")
	eventsCreateCmd.Flags().StringVar(&eventsCreateSummary, "summary", "", "event title (required)")
	eventsCreateCmd.Flags().StringVar(&eventsCreateStart, "start", "", "start: YYYY-MM-DDTHH:MM:SS or YYYY-MM-DD (required)")
	eventsCreateCmd.Flags().StringVar(&eventsCreateEnd, "end", "", "end: YYYY-MM-DDTHH:MM:SS or YYYY-MM-DD (required)")
	eventsCreateCmd.Flags().BoolVar(&eventsCreateAllDay, "allday", false, "mark as all-day event")
	eventsCreateCmd.Flags().StringVar(&eventsCreateLocation, "location", "", "event location")
	eventsCreateCmd.Flags().StringVar(&eventsCreateNotes, "notes", "", "event notes/description")
	eventsCreateCmd.Flags().StringVar(&eventsCreateURL, "url", "", "URL associated with the event")

	// update
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateCalendar, "calendar", "", "restrict search to this calendar (faster)")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateSummary, "summary", "", "new event title")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateStart, "start", "", "new start: YYYY-MM-DDTHH:MM:SS or YYYY-MM-DD")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateEnd, "end", "", "new end: YYYY-MM-DDTHH:MM:SS or YYYY-MM-DD")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateAllDay, "allday", "", "set all-day: true or false")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateLocation, "location", "", "new location")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateNotes, "notes", "", "new notes/description")
	eventsUpdateCmd.Flags().StringVar(&eventsUpdateURL, "url", "", "new URL")

	// delete
	eventsDeleteCmd.Flags().StringVar(&eventsDeleteCalendar, "calendar", "", "restrict search to this calendar (faster)")
	eventsDeleteCmd.Flags().BoolVar(&eventsDeleteConfirm, "confirm", false, "required: confirm permanent deletion")

	// open
	eventsOpenCmd.Flags().StringVar(&eventsOpenCalendar, "calendar", "", "restrict search to this calendar (faster)")

	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsGetCmd)
	eventsCmd.AddCommand(eventsCreateCmd)
	eventsCmd.AddCommand(eventsUpdateCmd)
	eventsCmd.AddCommand(eventsDeleteCmd)
	eventsCmd.AddCommand(eventsOpenCmd)

	rootCmd.AddCommand(eventsCmd)
}
