package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var downloadOutput string

var downloadCmd = &cobra.Command{
	Use:   "download <path>",
	Short: "Download a file",
	Long: `Download a file from your filesystem to the local machine.

The path argument must point to a file (not a folder).  By default the file
is saved in the current directory using its original name.  Use -o / --output
to specify a different destination path or file name.

Examples:
  fs download /readme.txt                    # save as ./readme.txt
  fs download /documents/report.pdf          # save as ./report.pdf
  fs download /docs/notes.txt -o ~/notes.txt # save to a custom location`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "destination file path (default: current directory, original file name)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	remotePath := args[0]

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
	file, err := apiClient.FindFileByPath(ctx, rootFileID, remotePath)
	if err != nil {
		return fmt.Errorf("file not found: %s (%v)", remotePath, err)
	}

	// ── Determine local destination ───────────────────────────────────────────
	dest := downloadOutput
	if dest == "" {
		dest = file.Name
	} else {
		// If the caller gave a directory, write the file inside it.
		info, statErr := os.Stat(dest)
		if statErr == nil && info.IsDir() {
			dest = filepath.Join(dest, file.Name)
		}
	}

	// ── Download ──────────────────────────────────────────────────────────────
	fmt.Printf("  Downloading %s …\n", file.Name)

	data, err := apiClient.DownloadDirect(ctx, file.ID)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// ── Save to disk ──────────────────────────────────────────────────────────
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("  Saved %d bytes → %s\n\n", len(data), dest)
	return nil
}
