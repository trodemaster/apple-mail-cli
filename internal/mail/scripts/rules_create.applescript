tell application "Mail"
	-- find the target mailbox
	set targetMailbox to missing value
	{{- if .AccountName}}
	try
		set targetMailbox to mailbox "{{.MoveMailbox}}" of account "{{.AccountName}}"
	on error
		set targetMailbox to missing value
	end try
	{{- else}}
	repeat with acc in accounts
		repeat with mbox in mailboxes of acc
			if name of mbox is "{{.MoveMailbox}}" then
				set targetMailbox to mbox
				exit repeat
			end if
		end repeat
		if targetMailbox is not missing value then exit repeat
	end repeat
	{{- end}}
	if targetMailbox is missing value then
		error "Mailbox not found: {{.MoveMailbox}}"
	end if

	-- create the rule
	set newRule to make new rule with properties {name:"{{.RuleName}}", enabled:true, all conditions must be met:false, stop evaluating rules:false}
	set should move message of newRule to true
	set move message of newRule to targetMailbox

	-- add the initial condition (set qualifier separately to avoid conflict with AppleScript's "ends with" operator)
	set newCond to make new rule condition at end of rule conditions of newRule with properties {rule type:{{.RuleTypeAS}}, expression:"{{.Expression}}"}
	set qualifier of newCond to {{.QualifierAS}}

	return "ok"
end tell
