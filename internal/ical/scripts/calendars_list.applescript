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

tell application "Calendar"
	set q to (ASCII character 34)
	set resultJSON to "["
	set cals to every calendar
	set calCount to count of cals
	repeat with i from 1 to calCount
		set cal to item i of cals
		set calName to name of cal
		set calID to ""
		try
			set calID to calendarIdentifier of cal
		on error
			set calID to calName
		end try
		set calWritable to writable of cal
		set calDesc to ""
		try
			set calDesc to description of cal
			if calDesc is missing value then set calDesc to ""
		end try

		if calWritable then
			set writableStr to "true"
		else
			set writableStr to "false"
		end if

		if i > 1 then set resultJSON to resultJSON & ","
		set resultJSON to resultJSON & "{" & q & "id" & q & ":" & q & my escapeJSON(calID) & q & "," & q & "name" & q & ":" & q & my escapeJSON(calName) & q & "," & q & "writable" & q & ":" & writableStr & "," & q & "description" & q & ":" & q & my escapeJSON(calDesc) & q & "}"
	end repeat
	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
