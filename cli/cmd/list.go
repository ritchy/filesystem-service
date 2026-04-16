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

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List files and folders",
	Long: `List files and folders in your filesystem.

If no path is given, the root directory is listed.

Examples:
  fs list              # list root directory
  fs list /            # list root directory
  fs list /documents   # list contents of the documents folder
  fs list /docs/work   # list a nested folder`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	path := "/"
	if len(args) > 0 {
		path = args[0]
	}

	// Load stored credentials
	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("not logged in – run 'fs login' first")
	}

	// Extract user ID (sub) from the JWT ID token
	userID, err := auth.GetUserIDFromToken(creds.IDToken)
	if err != nil {
		return fmt.Errorf("session expired – run 'fs login' again")
	}

	apiClient := api.NewClient(creds.IDToken)
	ctx := context.Background()

	// Resolve the user's root file ID via the Member → FileFolder relationship
	member, err := apiClient.GetMemberByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve account info: %w", err)
	}

	if member.FileFolder.ID == "" {
		return fmt.Errorf("no file system found for your account")
	}

	rootFileID := member.FileFolder.RootFileID

	// Navigate the path to find the target folder ID
	targetFolderID, err := apiClient.NavigatePath(ctx, rootFileID, path)
	if err != nil {
		return fmt.Errorf("path not found: %s (%v)", path, err)
	}

	// List the children of the target folder
	files, err := apiClient.ListFiles(ctx, targetFolderID)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// ── Output ────────────────────────────────────────────────────────────────
	displayPath := path
	if displayPath == "/" || displayPath == "" {
		displayPath = "/ (root)"
	}

	fmt.Printf("\n  Path: %s\n", displayPath)
	fmt.Println(strings.Repeat("─", 68))

	if len(files) == 0 {
		fmt.Println("  (empty)")
	} else {
		fmt.Printf("  %-40s  %-6s  %10s\n", "NAME", "TYPE", "SIZE")
		fmt.Println(strings.Repeat("─", 68))
		for _, file := range files {
			typeLabel := "file"
			sizeStr := "-"
			if file.Type == "folder" {
				typeLabel = "folder"
			} else if file.Size != nil {
				sizeStr = formatSize(*file.Size)
			}
			name := file.Name
			if file.Type == "folder" {
				name = name + "/"
			}
			fmt.Printf("  %-40s  %-6s  %10s\n", name, typeLabel, sizeStr)
		}
	}

	fmt.Printf("\n  %d item(s)\n\n", len(files))
	return nil
}

// formatSize converts a byte count to a human-readable string.
func formatSize(bytes int) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	case bytes < 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	default:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}
