tell application "Mail"
	set targetMailbox to mailbox "{{.MailboxName}}" of account "{{.AccountName}}"
	set beforeCount to count of messages of targetMailbox
	if beforeCount is 0 then return "0"
	set endIdx to {{.BatchSize}}
	if endIdx > beforeCount then set endIdx to beforeCount
	set batch to messages 1 through endIdx of targetMailbox
	perform mail action with messages batch
	set afterCount to count of messages of targetMailbox
	return (beforeCount - afterCount) as string
end tell
