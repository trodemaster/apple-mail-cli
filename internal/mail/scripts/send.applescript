tell application "Mail"
	set newMsg to make new outgoing message with properties {subject:"{{.Subject}}", content:"{{.Body}}", visible:false}

	tell newMsg
		{{range .To}}make new to recipient at end of to recipients with properties {address:"{{.}}"}
		{{end}}
		{{range .CC}}make new cc recipient at end of cc recipients with properties {address:"{{.}}"}
		{{end}}
		{{range .BCC}}make new bcc recipient at end of bcc recipients with properties {address:"{{.}}"}
		{{end}}
		{{range .Attachments}}make new attachment with properties {file name:(POSIX file "{{.}}")} at after last paragraph
		{{end}}
	end tell

	send newMsg
end tell
return "{\"sent\":true}"
