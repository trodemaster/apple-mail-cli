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
	set allRules to every rule
	set ruleCount to count of allRules
	set resultJSON to "["
	set isFirst to true

	repeat with i from 1 to ruleCount
		set theRule to item i of allRules
		set ruleName to name of theRule
		set ruleEnabled to (enabled of theRule) as string
		set ruleAllMust to (all conditions must be met of theRule) as string
		set ruleStop to (stop evaluating rules of theRule) as string

		-- move action: only dereference the mailbox when the flag is true
		set moveMailboxName to ""
		if should move message of theRule then
			try
				set moveMailboxName to my escapeJSON(name of (move message of theRule))
			on error
				set moveMailboxName to ""
			end try
		end if

		-- copy action
		set copyMailboxName to ""
		if should copy message of theRule then
			try
				set copyMailboxName to my escapeJSON(name of (copy message of theRule))
			on error
				set copyMailboxName to ""
			end try
		end if

		-- scalar boolean actions (coerce to string, emitted without surrounding quotes in JSON)
		set ruleDelete to (delete message of theRule) as string
		set ruleMarkRead to (mark read of theRule) as string
		set ruleMarkFlagged to (mark flagged of theRule) as string

		-- text actions (always readable; may be empty string)
		set ruleForward to my escapeJSON(forward message of theRule)
		set ruleRedirect to my escapeJSON(redirect message of theRule)

		-- build actions JSON object (booleans without surrounding q)
		set actionsJSON to "{" & q & "moveToMailbox" & q & ":" & q & moveMailboxName & q & "," & q & "copyToMailbox" & q & ":" & q & copyMailboxName & q & "," & q & "deleteMessage" & q & ":" & ruleDelete & "," & q & "markRead" & q & ":" & ruleMarkRead & "," & q & "markFlagged" & q & ":" & ruleMarkFlagged & "," & q & "forwardTo" & q & ":" & q & ruleForward & q & "," & q & "redirectTo" & q & ":" & q & ruleRedirect & q & "}"

		-- build conditions JSON array
		set condsJSON to "["
		set isFirstCond to true
		set theConds to rule conditions of theRule
		set condCount to count of theConds
		repeat with j from 1 to condCount
			set theCond to item j of theConds
			set condType to my escapeJSON((rule type of theCond) as string)
			set condQual to my escapeJSON((qualifier of theCond) as string)
			set condExpr to my escapeJSON(expression of theCond)
			set condHeader to ""
			try
				set condHeader to my escapeJSON(header of theCond)
			on error
				set condHeader to ""
			end try

			if not isFirstCond then set condsJSON to condsJSON & ","
			set isFirstCond to false
			set condsJSON to condsJSON & "{" & q & "ruleType" & q & ":" & q & condType & q & "," & q & "qualifier" & q & ":" & q & condQual & q & "," & q & "expression" & q & ":" & q & condExpr & q & "," & q & "header" & q & ":" & q & condHeader & q & "}"
		end repeat
		set condsJSON to condsJSON & "]"

		-- append rule to result array (boolean fields without surrounding q)
		if not isFirst then set resultJSON to resultJSON & ","
		set isFirst to false
		set resultJSON to resultJSON & "{" & q & "name" & q & ":" & q & my escapeJSON(ruleName) & q & "," & q & "enabled" & q & ":" & ruleEnabled & "," & q & "allConditionsMustBeMet" & q & ":" & ruleAllMust & "," & q & "stopEvaluatingRules" & q & ":" & ruleStop & "," & q & "conditions" & q & ":" & condsJSON & "," & q & "actions" & q & ":" & actionsJSON & "}"
	end repeat

	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
