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

on shouldSkipMailbox(mboxName)
	-- skip mailboxes that are unlikely to contain relevant received mail
	set skipList to {"Trash", "Deleted Messages", "Junk", "Spam", "Sent Messages", "Sent", "Drafts", "[Gmail]/Spam", "[Gmail]/Trash", "[Gmail]/Sent Mail", "[Airmail]", "All Mail"}
	repeat with skipName in skipList
		if mboxName is equal to skipName then return true
	end repeat
	return false
end shouldSkipMailbox

tell application "Mail"
	set q to (ASCII character 34)
	set queryStr to "{{.Query}}"
	set filterAccount to "{{.AccountName}}"
	set fromFilter to "{{.From}}"
	set subjectFilter to "{{.Subject}}"
	set mailboxFilter to "{{.MailboxName}}"
	set msgLimit to {{.Limit}}

	set resultJSON to "["
	set isFirst to true
	set totalCount to 0

	repeat with acc in accounts
		set accName to name of acc
		if filterAccount is "" or filterAccount is equal to accName then
			repeat with mbox in mailboxes of acc
				set mboxName to name of mbox

				if mailboxFilter is not "" and mboxName is not equal to mailboxFilter then
					-- skip: not the requested mailbox
				else if mailboxFilter is "" and my shouldSkipMailbox(mboxName) then
					-- skip: noise mailbox in open search
				else
					-- Use Mail.app's indexed 'whose' filtering for field-specific criteria.
					-- This is orders of magnitude faster than iterating all messages.
					if fromFilter is not "" and subjectFilter is not "" then
						set candidateMsgs to (messages of mbox whose sender contains fromFilter and subject contains subjectFilter)
					else if fromFilter is not "" then
						set candidateMsgs to (messages of mbox whose sender contains fromFilter)
					else if subjectFilter is not "" then
						set candidateMsgs to (messages of mbox whose subject contains subjectFilter)
					else
						set candidateMsgs to messages of mbox
					end if

					repeat with msg in candidateMsgs
						if msgLimit > 0 and totalCount >= msgLimit then exit repeat
						try
							-- queryStr searches subject+sender+body (slow; scope with --mailbox when using)
							set passesQuery to true
							if queryStr is not "" then
								if subject of msg contains queryStr then
									set passesQuery to true
								else if sender of msg contains queryStr then
									set passesQuery to true
								else if content of msg contains queryStr then
									set passesQuery to true
								else
									set passesQuery to false
								end if
							end if

							if passesQuery then
								set msgID to message id of msg
								set msgSubject to subject of msg
								set msgSender to sender of msg
								set msgDate to date sent of msg as string
								set isRead to read status of msg
								set msgFlagged to flagged status of msg
								set msgHasAtt to (count of mail attachments of msg) > 0
								set msgBody to ""
								{{if .IncludeBody}}set msgBody to my escapeJSON(content of msg){{end}}

								if not isFirst then set resultJSON to resultJSON & ","
								set isFirst to false

								set resultJSON to resultJSON & "{" & q & "id" & q & ":" & q & my escapeJSON(msgID) & q & "," & q & "subject" & q & ":" & q & my escapeJSON(msgSubject) & q & "," & q & "sender" & q & ":" & q & my escapeJSON(msgSender) & q & "," & q & "date" & q & ":" & q & msgDate & q & "," & q & "read" & q & ":" & (isRead as string) & "," & q & "flagged" & q & ":" & (msgFlagged as string) & "," & q & "hasAttachments" & q & ":" & (msgHasAtt as string) & "," & q & "body" & q & ":" & q & msgBody & q & "}"
								set totalCount to totalCount + 1
							end if
						on error
							-- skip messages that cannot be introspected
						end try
					end repeat
				end if
				if msgLimit > 0 and totalCount >= msgLimit then exit repeat
			end repeat
		end if
		if msgLimit > 0 and totalCount >= msgLimit then exit repeat
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
