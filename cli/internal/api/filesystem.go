package api

import (
	"context"
	"fmt"
	"strings"
)

// ── Domain types ──────────────────────────────────────────────────────────────

// FileItem represents a file or folder node in the filesystem tree.
type FileItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"` // "file" | "folder"
	Size            *int   `json:"size"`
	CreatedDate     string `json:"createdDate"`
	LastUpdatedDate string `json:"lastUpdatedDate"`
	ParentFileID    string `json:"parentFileId"`
	FileFolderID    string `json:"fileFolderId"`
}

// FileInfo contains metadata returned by the /info/:id REST endpoint.
type FileInfo struct {
	Count string `json:"count"`
	Size  string `json:"size"`
}

// FileFolder is the root container belonging to a Member.
type FileFolder struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	RootFileID string `json:"rootFileId"`
}

// Member is the user account record stored in the database.
type Member struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	Name       string     `json:"name"`
	Handle     string     `json:"handle"`
	FileFolder FileFolder `json:"fileFolder"`
}

// ── API methods ───────────────────────────────────────────────────────────────

// GetMemberByUserID fetches the Member whose userId matches the Cognito sub.
func (c *Client) GetMemberByUserID(ctx context.Context, userID string) (*Member, error) {
	const query = `
	query ListMembers($filter: ModelMemberFilterInput) {
		listMembers(filter: $filter) {
			items {
				id
				userId
				name
				handle
				fileFolder {
					id
					name
					rootFileId
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"userId": map[string]interface{}{
				"eq": userID,
			},
		},
	}

	var result struct {
		ListMembers struct {
			Items []Member `json:"items"`
		} `json:"listMembers"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("listMembers query: %w", err)
	}

	if len(result.ListMembers.Items) == 0 {
		return nil, fmt.Errorf("no account found for user ID %q", userID)
	}

	return &result.ListMembers.Items[0], nil
}

// ListFiles returns the immediate children of the folder identified by parentFileID.
func (c *Client) ListFiles(ctx context.Context, parentFileID string) ([]FileItem, error) {
	const query = `
	query ListFiles($filter: ModelFileFilterInput) {
		listFiles(filter: $filter) {
			items {
				id
				name
				type
				size
				createdDate
				lastUpdatedDate
				parentFileId
				fileFolderId
			}
		}
	}`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"parentFileId": map[string]interface{}{
				"eq": parentFileID,
			},
		},
	}

	var result struct {
		ListFiles struct {
			Items []FileItem `json:"items"`
		} `json:"listFiles"`
	}

	if err := c.execute(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("listFiles query: %w", err)
	}

	return result.ListFiles.Items, nil
}

// FindFileByPath resolves a full slash-separated path to a file (not a folder)
// and returns its FileItem.  The last segment of the path must be a file; all
// preceding segments must be folders.
//
// Examples:
//
//	FindFileByPath(ctx, rootID, "/readme.txt")          → FileItem for readme.txt
//	FindFileByPath(ctx, rootID, "/documents/report.pdf") → FileItem for report.pdf
func (c *Client) FindFileByPath(ctx context.Context, rootFileID string, path string) (*FileItem, error) {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	if path == "" {
		return nil, fmt.Errorf("path must point to a file, not the root directory")
	}

	parts := strings.Split(path, "/")
	fileName := parts[len(parts)-1]
	parentParts := parts[:len(parts)-1]

	// Resolve the parent folder.
	parentFolderID := rootFileID
	if len(parentParts) > 0 {
		parentPath := strings.Join(parentParts, "/")
		var err error
		parentFolderID, err = c.NavigatePath(ctx, rootFileID, parentPath)
		if err != nil {
			return nil, fmt.Errorf("parent path not found: %w", err)
		}
	}

	// List the parent folder's children and find the file by name.
	children, err := c.ListFiles(ctx, parentFolderID)
	if err != nil {
		return nil, fmt.Errorf("listing folder contents: %w", err)
	}

	for _, f := range children {
		if strings.EqualFold(f.Name, fileName) && f.Type == "file" {
			item := f
			return &item, nil
		}
	}

	return nil, fmt.Errorf("file '%s' not found", fileName)
}

// NavigatePath resolves a slash-separated path starting from rootFileID and
// returns the ID of the folder at the end of the path.
//
// Examples:
//
//	NavigatePath(ctx, rootID, "/")           → rootID (unchanged)
//	NavigatePath(ctx, rootID, "/documents")  → ID of the "documents" folder
//	NavigatePath(ctx, rootID, "/docs/work")  → ID of "work" inside "docs"
func (c *Client) NavigatePath(ctx context.Context, rootFileID string, path string) (string, error) {
	// Normalize: strip leading/trailing slashes and spaces.
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")

	// Root path → return rootFileID directly.
	if path == "" {
		return rootFileID, nil
	}

	parts := strings.Split(path, "/")
	currentID := rootFileID

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		children, err := c.ListFiles(ctx, currentID)
		if err != nil {
			return "", fmt.Errorf("listing '%s': %w", part, err)
		}

		found := false
		for _, f := range children {
			if strings.EqualFold(f.Name, part) && f.Type == "folder" {
				currentID = f.ID
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("folder '%s' not found", part)
		}
	}

	return currentID, nil
}
