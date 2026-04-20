package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the local metadata cache",
	Long: `Download your entire filesystem metadata by walking the File tree
from your root FileFolder and cache it to ~/.filesystem/metadata.json.

The command queries the AppSync GraphQL API directly: it looks up the
authenticated user's Member record, reads the rootFileId from the associated
FileFolder, then recursively walks the File tree using listFiles/getFile
queries. The cached JSON is used to power shell tab-completion of paths for
the other commands (list, move, rename, share, upload, create, delete,
download). Re-run 'fs refresh' after adding or removing files to keep
completion up to date.`,
	Args: cobra.NoArgs,
	RunE: runRefresh,
}

func runRefresh(cmd *cobra.Command, args []string) error {
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

	fmt.Println("  Downloading metadata…")

	// ── Resolve user root ─────────────────────────────────────────────────────
	fmt.Println("  • Looking up account…")
	member, err := apiClient.GetMemberByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve account info: %w", err)
	}
	if member.FileFolder.ID == "" || member.FileFolder.RootFileID == "" {
		return fmt.Errorf("no file system found for your account")
	}

	// ── Walk the File tree ────────────────────────────────────────────────────
	fmt.Println("  • Walking file tree…")
	var stats refreshStats
	root, err := buildTree(ctx, apiClient, member.FileFolder.RootFileID, &stats)
	if err != nil {
		return fmt.Errorf("failed to walk file tree: %w", err)
	}

	// ── Build the metadata document ───────────────────────────────────────────
	md := config.Metadata{
		FileFolderID: member.FileFolder.ID,
		RootFileID:   member.FileFolder.RootFileID,
		MemberID:     member.ID,
		UserID:       userID,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
		Root:         root,
	}

	raw, err := json.MarshalIndent(&md, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}

	// ── Persist ───────────────────────────────────────────────────────────────
	if err := config.SaveMetadata(raw); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	path, _ := config.MetadataPath()
	fmt.Printf("  ✓ Saved metadata to %s\n", path)
	fmt.Printf("    %d folder(s), %d file(s)\n", stats.folders, stats.files)
	fmt.Printf("    Generated: %s\n", md.GeneratedAt)
	return nil
}

// refreshStats tracks totals as the tree is walked.
type refreshStats struct {
	files   int
	folders int
}

// buildTree fetches the File at rootID and recursively walks its descendants,
// returning a MetadataNode tree suitable for persisting as metadata.json.
//
// The root node is fetched via getFile (so its concrete type/name/timestamps
// are known); children are retrieved via listFiles(parentFileId = …) so only
// the minimum number of round-trips are made.
func buildTree(ctx context.Context, c *api.Client, rootID string, stats *refreshStats) (*config.MetadataNode, error) {
	item, err := c.GetFileByID(ctx, rootID)
	if err != nil {
		return nil, fmt.Errorf("fetch root file %s: %w", rootID, err)
	}
	node := nodeFromItem(item)

	if node.Type == "folder" {
		stats.folders++
		children, err := c.ListFiles(ctx, rootID)
		if err != nil {
			return nil, fmt.Errorf("list children of %s: %w", rootID, err)
		}
		for _, child := range children {
			childNode, err := walkChild(ctx, c, child, stats)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, childNode)
		}
	} else {
		stats.files++
	}

	return node, nil
}

// walkChild converts a FileItem returned by ListFiles into a MetadataNode,
// recursing into sub-folders as needed. Unlike buildTree, it does not fetch
// the item again – the listFiles result already contains everything we need.
func walkChild(ctx context.Context, c *api.Client, item api.FileItem, stats *refreshStats) (*config.MetadataNode, error) {
	node := nodeFromItem(&item)

	if item.Type != "folder" {
		stats.files++
		return node, nil
	}

	stats.folders++
	children, err := c.ListFiles(ctx, item.ID)
	if err != nil {
		return nil, fmt.Errorf("list children of %s: %w", item.ID, err)
	}
	for _, child := range children {
		childNode, err := walkChild(ctx, c, child, stats)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}
	return node, nil
}

// nodeFromItem converts an api.FileItem to a config.MetadataNode, copying
// the fields that are available. Size is dereferenced when present.
func nodeFromItem(item *api.FileItem) *config.MetadataNode {
	n := &config.MetadataNode{
		ID:              item.ID,
		Name:            item.Name,
		Type:            item.Type,
		CreatedDate:     item.CreatedDate,
		LastUpdatedDate: item.LastUpdatedDate,
		ParentFileID:    item.ParentFileID,
		FileFolderID:    item.FileFolderID,
	}
	if item.Size != nil {
		n.Size = *item.Size
	}
	return n
}
