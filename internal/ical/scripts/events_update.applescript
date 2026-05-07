on makeDateTime(y, m, d, secs)
	set theDate to current date
	set day of theDate to 1
	set year of theDate to y
	set month of theDate to m
	set day of theDate to d
	set time of theDate to secs
	return theDate
end makeDateTime

tell application "Calendar"
	set targetUID to {{asLiteral .UID}}
	set calFilter to {{asLiteral .CalendarName}}
	set targetEvt to missing value

	if calFilter is "" then
		set searchCals to every calendar
	else
		set searchCals to {}
		repeat with cal in every calendar
			if name of cal is calFilter then
				copy cal to end of searchCals
				exit repeat
			end if
		end repeat
	end if

	repeat with cal in searchCals
		try
			set found to (every event of cal whose uid is targetUID)
			if (count of found) > 0 then
				set targetEvt to item 1 of found
				exit repeat
			end if
		end try
	end repeat

	if targetEvt is missing value then
		error "event not found: " & targetUID
	end if

	{{if .SetSummary}}set summary of targetEvt to {{asLiteral .Summary}}
	{{end}}{{if .SetStart}}set start date of targetEvt to my makeDateTime({{.StartYear}}, {{.StartMonth}}, {{.StartDay}}, {{.StartSecs}})
	{{end}}{{if .SetEnd}}set end date of targetEvt to my makeDateTime({{.EndYear}}, {{.EndMonth}}, {{.EndDay}}, {{.EndSecs}})
	{{end}}{{if .SetAllDay}}set allday event of targetEvt to {{.AllDay}}
	{{end}}{{if .SetLocation}}set location of targetEvt to {{asLiteral .Location}}
	{{end}}{{if .SetNotes}}set description of targetEvt to {{asLiteral .Notes}}
	{{end}}{{if .SetURL}}set url of targetEvt to {{asLiteral .URL}}
	{{end}}{{if .SetStatus}}
	if "{{.Status}}" is "confirmed" then
		set status of targetEvt to confirmed
	else if "{{.Status}}" is "cancelled" then
		set status of targetEvt to cancelled
	else if "{{.Status}}" is "tentative" then
		set status of targetEvt to tentative
	else if "{{.Status}}" is "none" then
		set status of targetEvt to none
	end if
	{{end}}
	save

	set q to (ASCII character 34)
	return "{" & q & "uid" & q & ":" & q & uid of targetEvt & q & "}"
end tell
