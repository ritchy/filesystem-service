package cmd

import (
	"context"
	"fmt"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share <file-path>",
	Short: "Generate a shareable link for a file",
	Long: `Generate a pre-signed shareable URL for a file in your filesystem.

The link is time-limited (typically 3 hours) and can be shared with anyone.
Only files (not folders) can be shared.

Examples:
  fs share /documents/report.pdf
  fs share /photos/photo.png`,
	Args: cobra.ExactArgs(1),
	RunE: runShare,
}

func runShare(cmd *cobra.Command, args []string) error {
	filePath := args[0]

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

	// ── Locate the file ───────────────────────────────────────────────────────
	file, err := apiClient.FindFileByPath(ctx, rootFileID, filePath)
	if err != nil {
		return fmt.Errorf("file not found: %s (%v)", filePath, err)
	}

	// ── Request share link ────────────────────────────────────────────────────
	link, err := apiClient.GetShareLink(ctx, file.ID)
	if err != nil {
		return fmt.Errorf("failed to generate share link: %w", err)
	}

	// ── Output ────────────────────────────────────────────────────────────────
	fmt.Printf("\n  File:    %s\n", file.Name)
	fmt.Printf("  Expires: %s\n", formatDate(link.Expires))
	fmt.Printf("\n  %s\n\n", link.URL)
	return nil
}
