package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/trodemaster/apple-mail-cli/internal/ical"
	"github.com/trodemaster/apple-mail-cli/internal/output"
)

var (
	eventsCalendar string
	eventsFrom     string
	eventsTo       string
	eventsToday    bool
	eventsDays     int
	eventsLimit    int
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List calendar events",
	Long: `List events from Calendar.app within a date range.

Defaults to events starting today through the next 7 days across all calendars.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		var from, to time.Time

		switch {
		case eventsToday:
			from = today
			to = today
		case eventsFrom != "" || eventsTo != "":
			var err error
			if eventsFrom != "" {
				from, err = time.ParseInLocation("2006-01-02", eventsFrom, now.Location())
				if err != nil {
					output.PrintError("events", "invalidDate", fmt.Sprintf("--from: %s", err), prettyFlag)
					return nil
				}
			} else {
				from = today
			}
			if eventsTo != "" {
				to, err = time.ParseInLocation("2006-01-02", eventsTo, now.Location())
				if err != nil {
					output.PrintError("events", "invalidDate", fmt.Sprintf("--to: %s", err), prettyFlag)
					return nil
				}
			} else {
				to = from.AddDate(0, 0, 6)
			}
		default:
			from = today
			to = today.AddDate(0, 0, eventsDays-1)
		}

		events, err := ical.ListEvents(ical.ListEventsParams{
			CalendarName: eventsCalendar,
			From:         from,
			To:           to,
			Limit:        eventsLimit,
		})
		if err != nil {
			output.PrintError("events", "executionFailed", err.Error(), prettyFlag)
			return nil
		}

		if isJSON(cmd) {
			output.PrintJSON("events", events, prettyFlag)
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

// formatEventTime splits an ISO datetime string into a date and time component
// for table display. All-day events show an empty time.
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
	rootCmd.AddCommand(eventsCmd)
	eventsCmd.Flags().StringVar(&eventsCalendar, "calendar", "", "filter to a specific calendar name")
	eventsCmd.Flags().StringVar(&eventsFrom, "from", "", "start date YYYY-MM-DD (default: today)")
	eventsCmd.Flags().StringVar(&eventsTo, "to", "", "end date YYYY-MM-DD (default: from + 6 days)")
	eventsCmd.Flags().BoolVar(&eventsToday, "today", false, "show only today's events")
	eventsCmd.Flags().IntVar(&eventsDays, "days", 7, "number of days from today (default: 7)")
	eventsCmd.Flags().IntVar(&eventsLimit, "limit", 50, "max events to return (0 = unlimited)")
}
