package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:     "delete <path>",
	Aliases: []string{"rm"},
	Short:   "Delete a file or folder",
	Long: `Delete a file or folder from your filesystem.

When deleting a file, you will be asked to confirm before deletion.
When deleting a folder, each item inside it (recursively) is confirmed
individually before anything is deleted.

Use --force / -f to skip all confirmation prompts.

Examples:
  fs delete /documents/report.pdf         # delete a file (with confirmation)
  fs delete /old-folder                   # delete a folder and all contents
  fs delete /old-folder --force           # skip confirmations`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompts")
}

func runDelete(cmd *cobra.Command, args []string) error {
	targetPath := args[0]

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

	// ── Collect items to delete ───────────────────────────────────────────────
	// Items are ordered so that children appear before their parents (safe
	// deletion order: leaves first).
	var toDelete []api.FileItem

	folderID, folderErr := apiClient.NavigatePath(ctx, rootFileID, targetPath)
	if folderErr == nil {
		// Guard against deleting the root.
		if folderID == rootFileID {
			return fmt.Errorf("cannot delete the root directory")
		}

		// Recursively collect all descendants (children first).
		descendants, err := collectDescendants(ctx, apiClient, folderID)
		if err != nil {
			return fmt.Errorf("failed to list folder contents: %w", err)
		}
		toDelete = append(toDelete, descendants...)

		// Fetch the folder's own FileItem so we can display its name.
		folderItem, err := apiClient.GetFileByID(ctx, folderID)
		if err != nil {
			// Fall back to a minimal representation if the lookup fails.
			name := lastSegment(targetPath)
			folderItem = &api.FileItem{ID: folderID, Name: name, Type: "folder"}
		}
		toDelete = append(toDelete, *folderItem)
	} else {
		// Try as a file.
		fileItem, fileErr := apiClient.FindFileByPath(ctx, rootFileID, targetPath)
		if fileErr != nil {
			return fmt.Errorf("path not found: %s", targetPath)
		}
		toDelete = append(toDelete, *fileItem)
	}

	if len(toDelete) == 0 {
		fmt.Println("\n  Nothing to delete.\n")
		return nil
	}

	// ── Confirm deletions ─────────────────────────────────────────────────────
	var confirmedIDs []string

	fmt.Println()
	for _, item := range toDelete {
		label := item.Name
		if item.Type == "folder" {
			label = item.Name + "/"
		}

		if deleteForce || confirmPrompt(fmt.Sprintf("  Delete %s %q?", item.Type, label)) {
			confirmedIDs = append(confirmedIDs, item.ID)
		}
	}

	if len(confirmedIDs) == 0 {
		fmt.Println("\n  Nothing deleted.\n")
		return nil
	}

	// ── Delete confirmed items ────────────────────────────────────────────────
	if err := apiClient.DeleteFiles(ctx, confirmedIDs); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	fmt.Printf("\n  Deleted %d item(s).\n\n", len(confirmedIDs))
	return nil
}

// collectDescendants recursively lists all descendants of folderID in
// deletion-safe order: a folder's children appear before the folder itself.
func collectDescendants(ctx context.Context, apiClient *api.Client, folderID string) ([]api.FileItem, error) {
	children, err := apiClient.ListFiles(ctx, folderID)
	if err != nil {
		return nil, err
	}

	var items []api.FileItem
	for _, child := range children {
		if child.Type == "folder" {
			// Recurse into sub-folders first so their contents appear before them.
			subItems, err := collectDescendants(ctx, apiClient, child.ID)
			if err != nil {
				return nil, err
			}
			items = append(items, subItems...)
		}
		items = append(items, child)
	}
	return items, nil
}

// confirmPrompt prints prompt and reads a y/n answer from stdin.
// Returns true only if the user types "y" or "yes" (case-insensitive).
func confirmPrompt(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
