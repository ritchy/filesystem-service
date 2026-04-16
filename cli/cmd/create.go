package cmd

import (
	"context"
	"fmt"
	"strings"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create <path>",
	Aliases: []string{"mkdir"},
	Short:   "Create a new folder",
	Long: `Create a new folder at the specified path.

All parent directories in the path must already exist.

Examples:
  fs create /documents              # create a folder at the root level
  fs create /documents/work         # create a nested folder`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	fullPath := args[0]

	// Split the path into parent directory and new folder name.
	newName := lastSegment(fullPath)
	if newName == "" {
		return fmt.Errorf("invalid path: %q – cannot determine folder name", fullPath)
	}

	// Derive the parent path (everything before the last segment).
	parentPath := "/"
	trimmed := strings.TrimRight(fullPath, "/")
	if idx := strings.LastIndex(trimmed, "/"); idx > 0 {
		parentPath = trimmed[:idx]
	}

	// ── Auth ──────────────────────────────────────────────────────────────────
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("not logged in – run 'fs login' first")
	}

	userID, err := auth.GetUserIDFromToken(creds.IDToken)
	if err != nil {
		return fmt.Errorf("session expired – run 'fs login' again")
	}

	apiClient := api.NewClient(creds.IDToken)
	ctx := context.Background()

	// ── Resolve user root ─────────────────────────────────────────────────────
	member, err := apiClient.GetMemberByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve account info: %w", err)
	}

	if member.FileFolder.ID == "" {
		return fmt.Errorf("no file system found for your account")
	}

	fileFolderID := member.FileFolder.ID
	rootFileID := member.FileFolder.RootFileID

	// ── Navigate to the parent folder ─────────────────────────────────────────
	parentFolderID, err := apiClient.NavigatePath(ctx, rootFileID, parentPath)
	if err != nil {
		return fmt.Errorf("parent path not found: %s (%v)", parentPath, err)
	}

	// ── Create the folder ─────────────────────────────────────────────────────
	created, err := apiClient.CreateFolder(ctx, parentFolderID, fileFolderID, newName)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	fmt.Printf("\n  Created folder %q (id: %s)\n\n", created.Name, created.ID)
	return nil
}
