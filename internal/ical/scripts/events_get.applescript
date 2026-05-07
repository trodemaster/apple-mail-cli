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

tell application "Calendar"
	set targetUID to {{asLiteral .UID}}
	set calFilter to {{asLiteral .CalendarName}}
	set targetEvt to missing value
	set targetCalName to ""

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
				set targetCalName to name of cal
				exit repeat
			end if
		end try
	end repeat

	if targetEvt is missing value then
		error "event not found: " & targetUID
	end if

	set q to (ASCII character 34)
	set evtUID to uid of targetEvt
	set evtSummary to summary of targetEvt
	if evtSummary is missing value then set evtSummary to ""
	set evtStart to my dateToISO(start date of targetEvt)
	set evtEnd to my dateToISO(end date of targetEvt)
	set evtAllDay to allday event of targetEvt

	set evtLocation to ""
	try
		set evtLocation to location of targetEvt
		if evtLocation is missing value then set evtLocation to ""
	end try

	set evtNotes to ""
	try
		set evtNotes to description of targetEvt
		if evtNotes is missing value then set evtNotes to ""
	end try

	set evtStatus to ""
	try
		set evtStatus to (status of targetEvt) as string
	end try

	set evtURL to ""
	try
		set evtURL to url of targetEvt
		if evtURL is missing value then set evtURL to ""
	end try

	if evtAllDay then
		set allDayStr to "true"
	else
		set allDayStr to "false"
	end if

	-- Build attendees array
	set attendeesJSON to "["
	set isFirstAtt to true
	try
		repeat with att in every attendee of targetEvt
			set attName to ""
			try
				set attName to display name of att
				if attName is missing value then set attName to ""
			end try
			set attEmail to ""
			try
				set attEmail to email of att
				if attEmail is missing value then set attEmail to ""
			end try
			set attStatus to ""
			try
				set attStatus to (participation status of att) as string
			end try
			if not isFirstAtt then set attendeesJSON to attendeesJSON & ","
			set isFirstAtt to false
			set attendeesJSON to attendeesJSON & "{" & q & "name" & q & ":" & q & my escapeJSON(attName) & q & "," & q & "email" & q & ":" & q & my escapeJSON(attEmail) & q & "," & q & "status" & q & ":" & q & attStatus & q & "}"
		end repeat
	end try
	set attendeesJSON to attendeesJSON & "]"

	set resultJSON to "{" & q & "uid" & q & ":" & q & my escapeJSON(evtUID) & q & "," & q & "summary" & q & ":" & q & my escapeJSON(evtSummary) & q & "," & q & "start" & q & ":" & q & evtStart & q & "," & q & "end" & q & ":" & q & evtEnd & q & "," & q & "allDay" & q & ":" & allDayStr & "," & q & "location" & q & ":" & q & my escapeJSON(evtLocation) & q & "," & q & "notes" & q & ":" & q & my escapeJSON(evtNotes) & q & "," & q & "status" & q & ":" & q & evtStatus & q & "," & q & "url" & q & ":" & q & my escapeJSON(evtURL) & q & "," & q & "calendar" & q & ":" & q & my escapeJSON(targetCalName) & q & "," & q & "attendees" & q & ":" & attendeesJSON & "}"
	return resultJSON
end tell
