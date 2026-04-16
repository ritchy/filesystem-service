package cmd

import (
	"errors"
	"fmt"

	"filesystem-cli/internal/api"

	"github.com/spf13/cobra"
)

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

		fmt.Println("\n  Session expired. Please log in again.")
		fmt.Println()

		if loginErr := runLogin(cmd, nil); loginErr != nil {
			return fmt.Errorf("re-authentication failed: %w", loginErr)
		}

		fmt.Println()

		// Retry the original command with the fresh credentials.
		return fn(cmd, args)
	}
}
