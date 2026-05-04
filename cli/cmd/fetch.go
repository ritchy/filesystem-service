package cmd

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch <url> [remote-folder-path]",
	Short: "Download a file from a URL",
	Long: `Download a file from the provided URL to the current directory.

If a remote folder path is also supplied the downloaded file is automatically
uploaded to that folder in your filesystem.

The command inspects the HTTP response to determine whether the URL points to
a downloadable file.  If it does not (e.g. an HTML page), you will be
informed and nothing is saved.

Examples:
  fs fetch https://example.com/file.png                # save to ./file.png
  fs fetch https://example.com/file.png /documents      # save locally AND upload to /documents`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runFetch,
}

func runFetch(cmd *cobra.Command, args []string) error {
	rawURL := args[0]

	// ── Validate URL ─────────────────────────────────────────────────────────
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("invalid URL: %s", rawURL)
	}

	// ── HTTP HEAD to pre-check ───────────────────────────────────────────────
	if !JSONOutputEnabled {
		fmt.Printf("  Checking URL…\n")
	}
	headResp, err := http.Head(rawURL)
	if err != nil {
		return fmt.Errorf("failed to reach URL: %w", err)
	}
	headResp.Body.Close()

	if headResp.StatusCode < 200 || headResp.StatusCode >= 400 {
		return fmt.Errorf("server returned HTTP %d for %s", headResp.StatusCode, rawURL)
	}

	contentType := headResp.Header.Get("Content-Type")
	contentDisp := headResp.Header.Get("Content-Disposition")

	// Decide whether this looks like a downloadable file.
	if !isDownloadableResource(contentType, contentDisp) {
		if JSONOutputEnabled {
			printJSON("fetch", map[string]interface{}{
				"url":         rawURL,
				"downloaded":  false,
				"contentType": contentType,
				"message":     "URL does not point to a downloadable file",
			})
			return nil
		}
		fmt.Printf("  The URL does not point to a downloadable file (Content-Type: %s).\n", contentType)
		fmt.Println("  Nothing to do.")
		return nil
	}

	// ── Derive file name ─────────────────────────────────────────────────────
	fileName := fileNameFromResponse(contentDisp, parsedURL)
	if fileName == "" {
		return fmt.Errorf("could not determine a file name from the URL")
	}

	// ── Download ─────────────────────────────────────────────────────────────
	if !JSONOutputEnabled {
		fmt.Printf("  Downloading %s …\n", fileName)
	}
	getResp, err := http.Get(rawURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode < 200 || getResp.StatusCode >= 400 {
		return fmt.Errorf("download failed: server returned HTTP %d", getResp.StatusCode)
	}

	data, err := io.ReadAll(getResp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Double-check the actual body content type in case the HEAD response was different.
	actualCT := getResp.Header.Get("Content-Type")
	if actualCT != "" && !isDownloadableResource(actualCT, getResp.Header.Get("Content-Disposition")) {
		if JSONOutputEnabled {
			printJSON("fetch", map[string]interface{}{
				"url":         rawURL,
				"downloaded":  false,
				"contentType": actualCT,
				"message":     "URL does not point to a downloadable file",
			})
			return nil
		}
		fmt.Printf("  The URL does not point to a downloadable file (Content-Type: %s).\n", actualCT)
		fmt.Println("  Nothing to do.")
		return nil
	}

	// ── Save locally ─────────────────────────────────────────────────────────
	localDest := filepath.Join(".", fileName)
	if err := os.WriteFile(localDest, data, 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	if !JSONOutputEnabled {
		fmt.Printf("  Saved %d bytes → %s\n", len(data), localDest)
	}

	// ── Optional upload ──────────────────────────────────────────────────────
	if len(args) < 2 {
		if JSONOutputEnabled {
			printJSON("fetch", map[string]interface{}{
				"url":        rawURL,
				"downloaded": true,
				"fileName":   fileName,
				"localPath":  localDest,
				"bytes":      len(data),
				"uploaded":   false,
			})
			return nil
		}
		fmt.Println()
		return nil
	}

	remotePath := args[1]
	if !JSONOutputEnabled {
		fmt.Printf("\n  Uploading to %s …\n", remotePath)
	}

	uploadedID, err := uploadFetchedFile(fileName, data, remotePath)
	if err != nil {
		return err
	}

	if JSONOutputEnabled {
		printJSON("fetch", map[string]interface{}{
			"url":        rawURL,
			"downloaded": true,
			"fileName":   fileName,
			"localPath":  localDest,
			"bytes":      len(data),
			"uploaded":   true,
			"remotePath": remotePath,
			"id":         uploadedID,
		})
	}

	return nil
}

// uploadFetchedFile performs the same steps as `fs upload` but operates on
// in-memory data instead of reading from disk again.  Returns the created
// file ID on success.
func uploadFetchedFile(fileName string, data []byte, remotePath string) (string, error) {
	// ── Auth ──────────────────────────────────────────────────────────────────
	creds, err := config.LoadCredentials()
	if err != nil {
		return "", fmt.Errorf("not logged in – run 'fs login' first")
	}

	userID, err := auth.GetUserIDFromToken(creds.IDToken)
	if err != nil {
		return "", fmt.Errorf("session expired – run 'fs login' again")
	}

	apiClient := api.NewClient(creds.IDToken)
	ctx := context.Background()

	// ── Resolve user root ─────────────────────────────────────────────────────
	member, err := apiClient.GetMemberByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve account info: %w", err)
	}

	if member.FileFolder.ID == "" {
		return "", fmt.Errorf("no file system found for your account")
	}

	rootFileID := member.FileFolder.RootFileID
	fileFolderID := member.FileFolder.ID

	// ── Resolve destination folder ────────────────────────────────────────────
	parentFolderID, err := apiClient.NavigatePath(ctx, rootFileID, remotePath)
	if err != nil {
		return "", fmt.Errorf("destination folder not found: %s (%v)", remotePath, err)
	}

	// ── Get AWS credentials via Cognito Identity Pool ─────────────────────────
	if !JSONOutputEnabled {
		fmt.Printf("  Authenticating with AWS…\n")
	}
	awsCreds, err := auth.GetAWSCredentials(ctx, creds.IDToken)
	if err != nil {
		return "", fmt.Errorf("failed to obtain AWS credentials: %w", err)
	}

	s3UserID := awsCreds.IdentityID
	if s3UserID == "" {
		s3UserID = userID
	}

	// ── Build S3 path ─────────────────────────────────────────────────────────
	s3Key := fmt.Sprintf("files/%s/%d_%s", s3UserID, time.Now().UnixMilli(), fileName)

	// ── Upload to S3 ──────────────────────────────────────────────────────────
	if !JSONOutputEnabled {
		fmt.Printf("  Uploading %s (%s) → s3://%s/%s\n", fileName, formatSize(len(data)), s3Bucket, s3Key)
	}

	if err := putS3Object(ctx, awsCreds, s3Key, fileName, data); err != nil {
		return "", fmt.Errorf("S3 upload failed: %w", err)
	}

	// ── Create File entry in the database ─────────────────────────────────────
	created, err := apiClient.CreateFile(ctx, parentFolderID, fileFolderID, fileName, s3Key, len(data))
	if err != nil {
		return "", fmt.Errorf("failed to create file record: %w", err)
	}

	if !JSONOutputEnabled {
		fmt.Printf("\n  ✓ Uploaded %q → %s  (id: %s)\n\n", fileName, remotePath, created.ID)
	}

	return created.ID, nil
}

// isDownloadableResource returns true when the Content-Type / Content-Disposition
// headers suggest the response body is a binary or non-HTML file rather than
// a web page.
func isDownloadableResource(contentType, contentDisposition string) bool {
	// An explicit attachment disposition always means "downloadable file".
	if strings.Contains(strings.ToLower(contentDisposition), "attachment") {
		return true
	}

	// Strip parameters (e.g. "; charset=utf-8").
	ct := strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	ct = strings.ToLower(ct)

	// Reject common non-file MIME types.
	switch {
	case ct == "text/html",
		ct == "application/xhtml+xml":
		return false
	case ct == "":
		// No Content-Type at all – treat as downloadable (best effort).
		return true
	}

	return true
}

// fileNameFromResponse extracts a file name from the Content-Disposition
// header or falls back to the last path segment of the URL.
func fileNameFromResponse(contentDisposition string, u *url.URL) string {
	// Try Content-Disposition first.
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			if fn, ok := params["filename"]; ok && fn != "" {
				return sanitizeFileName(fn)
			}
		}
	}

	// Fall back to the URL path.
	base := path.Base(u.Path)
	if base == "" || base == "/" || base == "." {
		return ""
	}

	return sanitizeFileName(base)
}

// sanitizeFileName removes characters that are problematic on common file systems.
func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == '\x00' {
			return '_'
		}
		return r
	}, name)

	// Remove query strings / fragments that may have leaked in.
	if idx := strings.IndexAny(name, "?#"); idx >= 0 {
		name = name[:idx]
	}

	// Un-escape percent-encoded characters for readability.
	if unescaped, err := url.PathUnescape(name); err == nil {
		name = unescaped
	}

	return name
}

// init is intentionally left empty — the command is registered in root.go.
// This avoids duplicate init() registration patterns and keeps wiring centralised.
