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

var createParents bool

var createCmd = &cobra.Command{
	Use:     "create <path>",
	Aliases: []string{"mkdir"},
	Short:   "Create a new folder",
	Long: `Create a new folder at the specified path.

By default all parent directories in the path must already exist.
Use -p / --parents to create the entire directory tree in one command;
any segments that already exist are silently reused.

Examples:
  fs create /documents              # create a folder at the root level
  fs create /documents/work         # create a nested folder (parents must exist)
  fs create -p /a/b/c               # create /a, /a/b, and /a/b/c as needed`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

func init() {
	createCmd.Flags().BoolVarP(&createParents, "parents", "p", false, "create parent directories as needed")
}

func runCreate(cmd *cobra.Command, args []string) error {
	fullPath := args[0]

	if createParents {
		return runCreateParents(fullPath)
	}

	return runCreateSingle(fullPath)
}

// runCreateSingle creates a single folder; the parent must already exist.
func runCreateSingle(fullPath string) error {
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

	if JSONOutputEnabled {
		printJSON("create", map[string]interface{}{
			"name": created.Name,
			"id":   created.ID,
			"type": "folder",
			"path": fullPath,
		})
		return nil
	}

	fmt.Printf("\n  Created folder %q (id: %s)\n\n", created.Name, created.ID)
	return nil
}

// runCreateParents walks each segment of the path, reusing existing folders
// and creating missing ones.  Mirrors `mkdir -p` behaviour.
func runCreateParents(fullPath string) error {
	// Normalise: strip leading/trailing slashes and split into segments.
	trimmed := strings.Trim(strings.TrimSpace(fullPath), "/")
	if trimmed == "" {
		return fmt.Errorf("invalid path: %q – nothing to create", fullPath)
	}

	segments := strings.Split(trimmed, "/")

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

	// ── Walk / create each segment ────────────────────────────────────────────
	type folderResult struct {
		Name    string `json:"name"`
		ID      string `json:"id"`
		Created bool   `json:"created"`
	}

	currentID := rootFileID
	results := make([]folderResult, 0, len(segments))
	createdCount := 0

	for _, seg := range segments {
		if seg == "" || seg == "." {
			continue
		}

		// Check if a folder with this name already exists under currentID.
		children, err := apiClient.ListFiles(ctx, currentID)
		if err != nil {
			return fmt.Errorf("failed to list folder contents: %w", err)
		}

		var existingID string
		for _, child := range children {
			if strings.EqualFold(child.Name, seg) && child.Type == "folder" {
				existingID = child.ID
				break
			}
		}

		if existingID != "" {
			// Folder already exists – reuse it.
			results = append(results, folderResult{Name: seg, ID: existingID, Created: false})
			currentID = existingID
		} else {
			// Create the missing folder.
			created, err := apiClient.CreateFolder(ctx, currentID, fileFolderID, seg)
			if err != nil {
				return fmt.Errorf("failed to create folder %q: %w", seg, err)
			}
			results = append(results, folderResult{Name: seg, ID: created.ID, Created: true})
			createdCount++
			currentID = created.ID
		}
	}

	// ── Output ────────────────────────────────────────────────────────────────
	if JSONOutputEnabled {
		printJSON("create", map[string]interface{}{
			"path":         fullPath,
			"createdCount": createdCount,
			"folders":      results,
		})
		return nil
	}

	if createdCount == 0 {
		fmt.Printf("\n  All folders in %q already exist.\n\n", fullPath)
	} else {
		fmt.Println()
		for _, r := range results {
			if r.Created {
				fmt.Printf("  Created folder %q (id: %s)\n", r.Name, r.ID)
			} else {
				fmt.Printf("  Exists  folder %q (id: %s)\n", r.Name, r.ID)
			}
		}
		fmt.Printf("\n  %d folder(s) created.\n\n", createdCount)
	}

	return nil
}
