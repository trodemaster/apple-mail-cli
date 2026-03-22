tell application "Mail"
	set sourceMailbox to mailbox "{{.MailboxName}}" of account "{{.AccountName}}"

	-- Find the rule and its move destination.
	set theRule to missing value
	repeat with r in every rule
		if name of r is "{{.RuleName}}" then
			set theRule to r
			exit repeat
		end if
	end repeat
	if theRule is missing value then
		error "Rule not found: {{.RuleName}}"
	end if
	if not (should move message of theRule) then
		error "Rule has no move action: {{.RuleName}}"
	end if
	set destMailbox to move message of theRule

	set totalMoved to 0

	-- Apply each condition using indexed 'whose' queries.
	-- IMPORTANT: pass the specifier directly to `move` — do NOT store in a variable first.
	-- Storing a whose-filtered result as a list causes "Can't make ... into type specifier" (-1700).
	-- Mail.app's sender property is "Display Name <email@domain.com>"; for ends-with we append ">".
	repeat with cond in rule conditions of theRule
		set expr to expression of cond
		set qual to qualifier of cond
		set condType to rule type of cond

		try
			if condType is from header then
				if qual is ends with value then
					set n to count of (messages of sourceMailbox whose sender ends with (expr & ">"))
					if n > 0 then
						move (messages of sourceMailbox whose sender ends with (expr & ">")) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is does contain value then
					set n to count of (messages of sourceMailbox whose sender contains expr)
					if n > 0 then
						move (messages of sourceMailbox whose sender contains expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is begins with value then
					set n to count of (messages of sourceMailbox whose sender starts with expr)
					if n > 0 then
						move (messages of sourceMailbox whose sender starts with expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is equal to value then
					set n to count of (messages of sourceMailbox whose sender is expr)
					if n > 0 then
						move (messages of sourceMailbox whose sender is expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				end if
			else if condType is subject header then
				if qual is does contain value then
					set n to count of (messages of sourceMailbox whose subject contains expr)
					if n > 0 then
						move (messages of sourceMailbox whose subject contains expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is ends with value then
					set n to count of (messages of sourceMailbox whose subject ends with expr)
					if n > 0 then
						move (messages of sourceMailbox whose subject ends with expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is begins with value then
					set n to count of (messages of sourceMailbox whose subject starts with expr)
					if n > 0 then
						move (messages of sourceMailbox whose subject starts with expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				else if qual is equal to value then
					set n to count of (messages of sourceMailbox whose subject is expr)
					if n > 0 then
						move (messages of sourceMailbox whose subject is expr) to destMailbox
						set totalMoved to totalMoved + n
					end if
				end if
			end if
		on error
			-- skip conditions that produce errors (unsupported type, missing mailbox, etc.)
		end try
	end repeat

	return totalMoved as string
end tell
