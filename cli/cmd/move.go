package cmd

import (
	"context"
	"fmt"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <source-path> <destination-folder-path>",
	Short: "Move a file or folder to a different folder",
	Long: `Move a file or folder to a different folder by updating its parent.

The destination must be an existing folder – moving to a file is not allowed.
The item keeps its original name; only its location changes.

Examples:
  fs move /folder/sub-folder /different_folder   # move a folder
  fs move /folder/file.txt   /different_folder   # move a file`,
	Args: cobra.ExactArgs(2),
	RunE: runMove,
}

func runMove(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

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

	rootFileID := member.FileFolder.RootFileID

	// ── Resolve source item ───────────────────────────────────────────────────
	var itemID string
	var itemName string
	var itemType string

	// Try folder first, then file.
	folderID, folderErr := apiClient.NavigatePath(ctx, rootFileID, srcPath)
	if folderErr == nil {
		// Guard against moving the root.
		if folderID == rootFileID {
			return fmt.Errorf("cannot move the root directory")
		}
		itemID = folderID
		itemType = "folder"
		itemName = lastSegment(srcPath)
	} else {
		fileItem, fileErr := apiClient.FindFileByPath(ctx, rootFileID, srcPath)
		if fileErr != nil {
			return fmt.Errorf("source path not found: %s", srcPath)
		}
		itemID = fileItem.ID
		itemType = "file"
		itemName = fileItem.Name
	}

	// ── Resolve destination folder ────────────────────────────────────────────
	// NavigatePath only resolves folders, so this implicitly validates that
	// the destination is a folder and not a file.
	destFolderID, err := apiClient.NavigatePath(ctx, rootFileID, dstPath)
	if err != nil {
		return fmt.Errorf("destination folder not found: %s", dstPath)
	}

	// ── Move ──────────────────────────────────────────────────────────────────
	moved, err := apiClient.MoveFile(ctx, itemID, destFolderID)
	if err != nil {
		return fmt.Errorf("move failed: %w", err)
	}

	fmt.Printf("\n  Moved %s %q → %s\n\n", itemType, itemName, dstPath)
	_ = moved
	return nil
}
