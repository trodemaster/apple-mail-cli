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

tell application "Mail"
	set q to (ASCII character 34)
	set accountName to "{{.AccountName}}"
	set mailboxName to "{{.MailboxName}}"
	set msgLimit to {{.Limit}}
	set unreadOnly to {{.UnreadOnly}}

	set targetMailbox to mailbox mailboxName of account accountName
	set allMsgs to messages of targetMailbox

	set resultJSON to "["
	set isFirst to true
	set msgCount to 0

	repeat with msg in allMsgs
		if msgLimit > 0 and msgCount >= msgLimit then exit repeat

		set isRead to read status of msg
		if unreadOnly and isRead then
			-- skip read messages
		else
			set msgID to message id of msg
			set msgSubject to subject of msg
			set msgSender to sender of msg
			set msgDate to date sent of msg as string
			set msgFlagged to flagged status of msg
			set msgHasAtt to (count of mail attachments of msg) > 0
			set msgBody to ""
			{{if .IncludeBody}}set msgBody to my escapeJSON(content of msg){{end}}

			if not isFirst then set resultJSON to resultJSON & ","
			set isFirst to false

			set resultJSON to resultJSON & "{" & q & "id" & q & ":" & q & my escapeJSON(msgID) & q & "," & q & "subject" & q & ":" & q & my escapeJSON(msgSubject) & q & "," & q & "sender" & q & ":" & q & my escapeJSON(msgSender) & q & "," & q & "date" & q & ":" & q & msgDate & q & "," & q & "read" & q & ":" & (isRead as string) & "," & q & "flagged" & q & ":" & (msgFlagged as string) & "," & q & "hasAttachments" & q & ":" & (msgHasAtt as string) & "," & q & "body" & q & ":" & q & msgBody & q & "}"
			set msgCount to msgCount + 1
		end if
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
