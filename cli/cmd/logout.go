package cmd

import (
	"fmt"

	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of filesystem.io",
	Long: `Log out of filesystem.io by removing stored credentials.

This deletes ~/.filesystem/credentials.json and the cached metadata file.
You will need to run 'fs login' again before using authenticated commands.`,
	RunE: runLogout,
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := config.DeleteCredentials(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	// Also remove the cached metadata since it is user-specific.
	if err := config.DeleteMetadata(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	if JSONOutputEnabled {
		printJSON("logout", map[string]interface{}{
			"message": "Logged out successfully",
		})
		return nil
	}

	fmt.Println("✓ Logged out successfully.")
	return nil
}
