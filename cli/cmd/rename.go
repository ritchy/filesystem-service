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

var renameCmd = &cobra.Command{
	Use:   "rename <source-path> <new-path>",
	Short: "Rename a file or folder",
	Long: `Rename a file or folder in your filesystem.

The new name is taken from the last segment of <new-path>.
The item is renamed in place – moving to a different directory is not supported.

Examples:
  fs rename /docs/old.txt /docs/new.txt       # rename a file
  fs rename /old-folder   /new-folder         # rename a folder`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func runRename(cmd *cobra.Command, args []string) error {
	srcPath := args[0]
	dstPath := args[1]

	// Derive the new name from the last segment of the destination path.
	newName := lastSegment(dstPath)
	if newName == "" {
		return fmt.Errorf("invalid destination: %q – cannot determine new name", dstPath)
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

	rootFileID := member.FileFolder.RootFileID

	// ── Resolve the source item ───────────────────────────────────────────────
	// Guard against renaming the root directory.
	if strings.Trim(strings.TrimSpace(srcPath), "/") == "" {
		return fmt.Errorf("cannot rename the root directory")
	}

	var itemID string
	var itemType string
	var oldName string

	// Try to resolve as a folder first.
	folderID, folderErr := apiClient.NavigatePath(ctx, rootFileID, srcPath)
	if folderErr == nil {
		itemID = folderID
		itemType = "folder"
		oldName = lastSegment(srcPath)
	} else {
		// Fall back to resolving as a file.
		fileItem, fileErr := apiClient.FindFileByPath(ctx, rootFileID, srcPath)
		if fileErr != nil {
			return fmt.Errorf("path not found: %s", srcPath)
		}
		itemID = fileItem.ID
		itemType = "file"
		oldName = fileItem.Name
	}

	// ── Rename ────────────────────────────────────────────────────────────────
	if err := apiClient.RenameFile(ctx, itemID, newName); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}

	fmt.Printf("\n  Renamed %s %q → %q\n\n", itemType, oldName, newName)
	return nil
}

// lastSegment returns the final non-empty path segment (e.g. "new.txt" from
// "/docs/new.txt" or "new-folder" from "/new-folder/").
func lastSegment(p string) string {
	p = strings.TrimRight(p, "/")
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		p = p[idx+1:]
	}
	return strings.TrimSpace(p)
}
