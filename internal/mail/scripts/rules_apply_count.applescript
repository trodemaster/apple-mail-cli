tell application "Mail"
	set targetMailbox to mailbox "{{.MailboxName}}" of account "{{.AccountName}}"
	return (count of messages of targetMailbox) as string
end tell
