tell application "Mail"
	set q to (ASCII character 34)
	set batchSize to {{.BatchSize}}
	set totalMoved to 0
	set accountsProcessed to 0

	set resultJSON to "["
	set isFirst to true

	repeat with acc in accounts
		set accName to name of acc
		if "{{.AccountName}}" is "" or accName is "{{.AccountName}}" then
			try
				set targetMailbox to mailbox "{{.MailboxName}}" of acc
				set originalCount to count of messages of targetMailbox
				set batchCount to (originalCount + batchSize - 1) div batchSize
				set passesRun to 0

				-- Process in batches; as rules move messages out, indices shift down,
				-- so we always grab from the front until we've run enough passes.
				repeat with pass from 1 to batchCount
					set remaining to count of messages of targetMailbox
					if remaining is 0 then exit repeat
					set endIdx to batchSize
					if endIdx > remaining then set endIdx to remaining
					set batch to messages 1 through endIdx of targetMailbox
					perform mail action with messages batch
					set passesRun to passesRun + 1
				end repeat

				set finalCount to count of messages of targetMailbox
				set movedCount to originalCount - finalCount

				if not isFirst then set resultJSON to resultJSON & ","
				set isFirst to false
				set resultJSON to resultJSON & "{" & q & "account" & q & ":" & q & accName & q & "," & q & "mailbox" & q & ":" & q & "{{.MailboxName}}" & q & "," & q & "originalCount" & q & ":" & originalCount & "," & q & "movedCount" & q & ":" & movedCount & "," & q & "finalCount" & q & ":" & finalCount & "}"

				set accountsProcessed to accountsProcessed + 1
			on error errMsg
				-- account doesn't have this mailbox or other error; skip silently
			end try
		end if
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
