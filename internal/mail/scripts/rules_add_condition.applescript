tell application "Mail"
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
	set newCond to make new rule condition at end of rule conditions of theRule with properties {rule type:{{.RuleTypeAS}}, expression:"{{.Expression}}"}
	set qualifier of newCond to {{.QualifierAS}}
	return "ok"
end tell
