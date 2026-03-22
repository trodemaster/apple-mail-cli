tell application "Mail"
	set q to (ASCII character 34)
	set resultJSON to "["
	set accs to accounts
	set accCount to count of accs
	repeat with i from 1 to accCount
		set acc to item i of accs
		set accName to name of acc
		set accEmails to email addresses of acc
		set accType to (account type of acc) as string

		set emailJSON to "["
		set emailCount to count of accEmails
		repeat with j from 1 to emailCount
			set emailJSON to emailJSON & q & (item j of accEmails) & q
			if j < emailCount then set emailJSON to emailJSON & ","
		end repeat
		set emailJSON to emailJSON & "]"

		set resultJSON to resultJSON & "{" & q & "id" & q & ":" & q & accName & q & "," & q & "name" & q & ":" & q & accName & q & "," & q & "emailAddresses" & q & ":" & emailJSON & "," & q & "type" & q & ":" & q & accType & q & "}"
		if i < accCount then set resultJSON to resultJSON & ","
	end repeat
	set resultJSON to resultJSON & "]"
	return resultJSON
end tell
