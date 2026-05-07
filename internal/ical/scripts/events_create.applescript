on escapeJSON(s)
	set q to (ASCII character 34)
	set bs to (ASCII character 92)
	set resultStr to ""
	repeat with i from 1 to length of s
		set c to character i of s
		if c = q then
			set resultStr to resultStr & bs & q
		else if c = bs then
			set resultStr to resultStr & bs & bs
		else if c = (ASCII character 9) then
			set resultStr to resultStr & bs & "t"
		else if c = (ASCII character 10) then
			set resultStr to resultStr & bs & "n"
		else if c = (ASCII character 13) then
			set resultStr to resultStr & bs & "r"
		else
			set resultStr to resultStr & c
		end if
	end repeat
	return resultStr
end escapeJSON

on dateToISO(d)
	set y to year of d as string
	set mo to (month of d as integer)
	set dy to day of d
	set hr to hours of d
	set mn to minutes of d
	set sc to seconds of d
	set moStr to mo as string
	set dyStr to dy as string
	set hrStr to hr as string
	set mnStr to mn as string
	set scStr to sc as string
	if mo < 10 then set moStr to "0" & moStr
	if dy < 10 then set dyStr to "0" & dyStr
	if hr < 10 then set hrStr to "0" & hrStr
	if mn < 10 then set mnStr to "0" & mnStr
	if sc < 10 then set scStr to "0" & scStr
	return y & "-" & moStr & "-" & dyStr & "T" & hrStr & ":" & mnStr & ":" & scStr
end dateToISO

-- Build a date from integer components, locale-independently.
-- Set day to 1 first to avoid month-boundary overflow.
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
	set calName to {{asLiteral .CalendarName}}
	set targetCal to missing value
	repeat with cal in every calendar
		if name of cal is calName and writable of cal is true then
			set targetCal to cal
			exit repeat
		end if
	end repeat
	if targetCal is missing value then
		error "writable calendar not found: " & calName
	end if

	set startDate to my makeDateTime({{.StartYear}}, {{.StartMonth}}, {{.StartDay}}, {{.StartSecs}})
	set endDate to my makeDateTime({{.EndYear}}, {{.EndMonth}}, {{.EndDay}}, {{.EndSecs}})

	set newEvt to make new event at end of events of targetCal with properties {summary: {{asLiteral .Summary}}, start date: startDate, end date: endDate}
	set allday event of newEvt to {{.AllDay}}
	{{if .Location}}set location of newEvt to {{asLiteral .Location}}
	{{end}}{{if .Notes}}set description of newEvt to {{asLiteral .Notes}}
	{{end}}{{if .URL}}set url of newEvt to {{asLiteral .URL}}
	{{end}}
	save

	set q to (ASCII character 34)
	set evtUID to uid of newEvt
	set evtSummary to summary of newEvt
	if evtSummary is missing value then set evtSummary to ""
	set evtStart to my dateToISO(start date of newEvt)
	set evtEnd to my dateToISO(end date of newEvt)
	set evtAllDay to allday event of newEvt
	if evtAllDay then
		set allDayStr to "true"
	else
		set allDayStr to "false"
	end if

	return "{" & q & "uid" & q & ":" & q & my escapeJSON(evtUID) & q & "," & q & "summary" & q & ":" & q & my escapeJSON(evtSummary) & q & "," & q & "start" & q & ":" & q & evtStart & q & "," & q & "end" & q & ":" & q & evtEnd & q & "," & q & "allDay" & q & ":" & allDayStr & "," & q & "calendar" & q & ":" & q & my escapeJSON(calName) & q & "}"
end tell
