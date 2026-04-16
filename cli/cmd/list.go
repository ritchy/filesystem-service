package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list [path]",
	Aliases: []string{"ls"},
	Short:   "List files and folders",
	Long: `List files and folders in your filesystem.

If no path is given, the root directory is listed.
If the path points to a file, its details are displayed instead.

Examples:
  fs list                       # list root directory
  fs list /                     # list root directory
  fs list /documents            # list contents of the documents folder
  fs list /docs/work            # list a nested folder
  fs list /docs/readme.txt      # display info for a specific file`,
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

	// Try to navigate to the path as a folder first.
	targetFolderID, folderErr := apiClient.NavigatePath(ctx, rootFileID, path)
	if folderErr == nil {
		// Path resolved to a folder – list its children.
		return listFolder(apiClient, ctx, path, targetFolderID)
	}

	// NavigatePath failed – try resolving the path as a file.
	fileItem, fileErr := apiClient.FindFileByPath(ctx, rootFileID, path)
	if fileErr == nil {
		// Path resolved to a file – display its info.
		return showFileInfo(apiClient, ctx, path, fileItem)
	}

	// Neither a folder nor a file could be resolved.
	return fmt.Errorf("path not found: %s", path)
}

// listFolder prints the immediate children of a folder.
func listFolder(apiClient *api.Client, ctx context.Context, path string, folderID string) error {
	files, err := apiClient.ListFiles(ctx, folderID)
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

// showFileInfo displays detailed information about a single file of type 'file'.
func showFileInfo(apiClient *api.Client, ctx context.Context, path string, item *api.FileItem) error {
	// Fetch additional info from the REST /info/:id endpoint (mirrors
	// fetchFileInfo() in the React app's api.ts).
	info, infoErr := apiClient.GetFileInfo(ctx, item.ID)

	// Prefer the size string returned by the API; fall back to local formatting.
	sizeStr := "-"
	if infoErr == nil && info.Size != "" {
		sizeStr = info.Size
	} else if item.Size != nil {
		sizeStr = formatSize(*item.Size)
	}

	fmt.Printf("\n  Path: %s\n", path)
	fmt.Println(strings.Repeat("─", 68))
	fmt.Printf("  %-16s %s\n", "Name:", item.Name)
	fmt.Printf("  %-16s %s\n", "Type:", "file")
	fmt.Printf("  %-16s %s\n", "Size:", sizeStr)
	fmt.Printf("  %-16s %s\n", "Created:", formatDate(item.CreatedDate))
	fmt.Printf("  %-16s %s\n", "Last Updated:", formatDate(item.LastUpdatedDate))
	fmt.Println()
	return nil
}

// formatDate parses an ISO 8601 timestamp and returns a readable UTC string.
func formatDate(iso string) string {
	if iso == "" {
		return "-"
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t.UTC().Format("2006-01-02 15:04:05 UTC")
		}
	}
	return iso // return raw value if no layout matched
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
