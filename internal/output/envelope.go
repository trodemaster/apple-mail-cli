package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Command   string `json:"command"`
	Timestamp string `json:"timestamp"`
}

type Envelope struct {
	OK    bool         `json:"ok"`
	Data  interface{}  `json:"data"`
	Error *ErrorDetail `json:"error"`
	Meta  Meta         `json:"meta"`
}

func PrintJSON(cmd string, data interface{}, pretty bool) {
	env := Envelope{
		OK:   true,
		Data: data,
		Meta: Meta{
			Command:   cmd,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	printEnvelope(env, pretty)
}

func PrintError(cmd string, code string, message string, pretty bool) {
	env := Envelope{
		OK:   false,
		Data: nil,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
		Meta: Meta{
			Command:   cmd,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	printEnvelope(env, pretty)
	os.Exit(1)
}

func printEnvelope(env Envelope, pretty bool) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(env, "", "  ")
	} else {
		b, err = json.Marshal(env)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}
