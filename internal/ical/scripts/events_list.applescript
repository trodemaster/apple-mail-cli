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

-- Build a date from integer year/month/day components, locale-independently.
-- Set day to 1 first to avoid month-boundary overflow when changing month.
on makeDate(y, m, d)
	set theDate to current date
	set day of theDate to 1
	set year of theDate to y
	set month of theDate to m
	set day of theDate to d
	set time of theDate to 0
	return theDate
end makeDate

tell application "Calendar"
	set q to (ASCII character 34)
	set calFilter to "{{.CalendarName}}"
	set limitCount to {{.Limit}}

	set fromDate to my makeDate({{.FromYear}}, {{.FromMonth}}, {{.FromDay}})
	set toDate to my makeDate({{.ToYear}}, {{.ToMonth}}, {{.ToDay}})
	set time of toDate to 86399 -- end of day

	set targetCals to {}
	if calFilter is "" then
		set targetCals to every calendar
	else
		try
			copy (first calendar whose name is calFilter) to end of targetCals
		on error
			return "[]"
		end try
	end if

	set resultJSON to "["
	set isFirst to true
	set totalCount to 0

	repeat with cal in targetCals
		set calName to name of cal
		set evts to {}
		try
			set evts to (every event of cal whose start date >= fromDate and start date <= toDate)
		end try

		repeat with evt in evts
			if limitCount > 0 and totalCount >= limitCount then exit repeat

			set evtUID to uid of evt
			set evtSummary to summary of evt
			if evtSummary is missing value then set evtSummary to ""

			set evtStart to my dateToISO(start date of evt)
			set evtEnd to my dateToISO(end date of evt)
			set evtAllDay to allday event of evt

			set evtLocation to ""
			try
				set evtLocation to location of evt
				if evtLocation is missing value then set evtLocation to ""
			end try

			set evtNotes to ""
			try
				set evtNotes to description of evt
				if evtNotes is missing value then set evtNotes to ""
			end try

			set evtStatus to ""
			try
				set evtStatus to (status of evt) as string
			end try

			set evtURL to ""
			try
				set evtURL to url of evt
				if evtURL is missing value then set evtURL to ""
			end try

			if evtAllDay then
				set allDayStr to "true"
			else
				set allDayStr to "false"
			end if

			if not isFirst then set resultJSON to resultJSON & ","
			set isFirst to false

			set resultJSON to resultJSON & "{" & q & "uid" & q & ":" & q & my escapeJSON(evtUID) & q & "," & q & "summary" & q & ":" & q & my escapeJSON(evtSummary) & q & "," & q & "start" & q & ":" & q & evtStart & q & "," & q & "end" & q & ":" & q & evtEnd & q & "," & q & "allDay" & q & ":" & allDayStr & "," & q & "location" & q & ":" & q & my escapeJSON(evtLocation) & q & "," & q & "notes" & q & ":" & q & my escapeJSON(evtNotes) & q & "," & q & "status" & q & ":" & q & evtStatus & q & "," & q & "url" & q & ":" & q & my escapeJSON(evtURL) & q & "," & q & "calendar" & q & ":" & q & my escapeJSON(calName) & q & "}"
			set totalCount to totalCount + 1
		end repeat

		if limitCount > 0 and totalCount >= limitCount then exit repeat
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
