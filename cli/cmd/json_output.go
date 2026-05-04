package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSONOutputEnabled is the global flag toggled by --json.
var JSONOutputEnabled bool

// JSONResponse is the standard envelope for all JSON output.
type JSONResponse struct {
	Success bool        `json:"success"`
	Command string      `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// printJSON marshals the given data as the JSON response and writes it to
// stdout. It should be the ONLY output when --json is active.
func printJSON(command string, data interface{}) {
	resp := JSONResponse{
		Success: true,
		Command: command,
		Data:    data,
	}
	raw, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		// Fallback: if we can't marshal, emit a minimal error object.
		fmt.Fprintf(os.Stdout, `{"success":false,"error":"failed to marshal JSON response: %s"}`+"\n", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(raw))
}

// printJSONError emits a JSON error object to stdout. When --json is active
// this replaces the normal stderr error output.
func printJSONError(command string, err error) {
	resp := JSONResponse{
		Success: false,
		Command: command,
		Error:   err.Error(),
	}
	raw, marshalErr := json.MarshalIndent(resp, "", "  ")
	if marshalErr != nil {
		fmt.Fprintf(os.Stdout, `{"success":false,"error":"%s"}`+"\n", err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, string(raw))
}
