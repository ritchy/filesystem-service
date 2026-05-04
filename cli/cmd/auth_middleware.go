package cmd

import (
	"errors"
	"fmt"

	"filesystem-cli/internal/api"

	"github.com/spf13/cobra"
)

// suppressLoginJSON is set to true during automatic re-authentication so
// that the login command does not emit its own JSON output.  In JSON mode
// the caller's command is the only thing that should produce output.
var suppressLoginJSON bool

// withAutoLogin wraps a cobra RunE function so that an expired or invalid
// session is handled transparently:
//
//  1. The original command runs.
//  2. If the error contains api.ErrUnauthorized (HTTP 401), the login flow
//     is triggered automatically (same as running 'fs login').
//  3. After a successful re-authentication the original command is retried
//     once with the freshly stored credentials.
func withAutoLogin(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := fn(cmd, args)
		if err == nil {
			return nil
		}

		// Only intercept unauthorized errors.
		if !errors.Is(err, api.ErrUnauthorized) {
			return err
		}

		if !JSONOutputEnabled {
			//fmt.Println("\n  Session expired. Please log in again.")
			fmt.Println()
		}

		// Suppress login JSON output during automatic re-auth.
		suppressLoginJSON = true
		loginErr := runLogin(cmd, nil)
		suppressLoginJSON = false

		if loginErr != nil {
			return fmt.Errorf("re-authentication failed: %w", loginErr)
		}

		if !JSONOutputEnabled {
			fmt.Println()
		}

		// Retry the original command with the fresh credentials.
		return fn(cmd, args)
	}
}
