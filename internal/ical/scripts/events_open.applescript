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

	show targetEvt
	activate
	return "true"
end tell
