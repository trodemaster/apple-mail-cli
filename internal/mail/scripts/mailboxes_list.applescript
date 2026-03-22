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
		else
			set resultStr to resultStr & c
		end if
	end repeat
	return resultStr
end escapeJSON

tell application "Mail"
	set q to (ASCII character 34)
	set resultJSON to "["
	set isFirst to true

	set filterAccount to "{{.AccountName}}"

	set accs to accounts
	repeat with acc in accs
		set accName to name of acc
		if filterAccount is "" or filterAccount is equal to accName then
			set mboxes to mailboxes of acc
			repeat with mbox in mboxes
				set mboxName to name of mbox
				set unread to unread count of mbox
				set total to count of messages of mbox

				if not isFirst then set resultJSON to resultJSON & ","
				set isFirst to false

				set resultJSON to resultJSON & "{" & q & "name" & q & ":" & q & my escapeJSON(mboxName) & q & "," & q & "account" & q & ":" & q & my escapeJSON(accName) & q & "," & q & "unreadCount" & q & ":" & unread & "," & q & "totalCount" & q & ":" & total & "}"
			end repeat
		end if
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
