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
		else if c = (ASCII character 10) then
			set resultStr to resultStr & bs & "n"
		else if c = (ASCII character 13) then
			set resultStr to resultStr & bs & "r"
		else if c = (ASCII character 9) then
			set resultStr to resultStr & bs & "t"
		else
			set resultStr to resultStr & c
		end if
	end repeat
	return resultStr
end escapeJSON

tell application "Mail"
	set q to (ASCII character 34)
	set targetID to "{{.MessageID}}"

	set foundMsg to missing value
	repeat with acc in accounts
		repeat with mbox in mailboxes of acc
			set msgs to (messages of mbox whose message id is targetID)
			if (count of msgs) > 0 then
				set foundMsg to item 1 of msgs
				exit repeat
			end if
		end repeat
		if foundMsg is not missing value then exit repeat
	end repeat

	if foundMsg is missing value then
		return "{" & q & "error" & q & ":" & q & "message not found" & q & "}"
	end if

	set msgID to message id of foundMsg
	set msgSubject to subject of foundMsg
	set msgSender to sender of foundMsg
	set msgDate to date sent of foundMsg as string
	set isRead to read status of foundMsg
	set msgFlagged to flagged status of foundMsg
	set msgContent to content of foundMsg

	set atts to mail attachments of foundMsg
	set attJSON to "["
	set attCount to count of atts
	repeat with k from 1 to attCount
		set att to item k of atts
		set attName to name of att
		set attJSON to attJSON & "{" & q & "name" & q & ":" & q & my escapeJSON(attName) & q & "}"
		if k < attCount then set attJSON to attJSON & ","
	end repeat
	set attJSON to attJSON & "]"

	return "{" & q & "id" & q & ":" & q & my escapeJSON(msgID) & q & "," & q & "subject" & q & ":" & q & my escapeJSON(msgSubject) & q & "," & q & "sender" & q & ":" & q & my escapeJSON(msgSender) & q & "," & q & "date" & q & ":" & q & msgDate & q & "," & q & "read" & q & ":" & (isRead as string) & "," & q & "flagged" & q & ":" & (msgFlagged as string) & "," & q & "body" & q & ":" & q & my escapeJSON(msgContent) & q & "," & q & "attachments" & q & ":" & attJSON & "}"
end tell
