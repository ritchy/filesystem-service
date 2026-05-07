package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MetadataNode mirrors the JSON returned by GET /files/metadata.
// Folders include a populated Children slice; files have Children == nil.
type MetadataNode struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Type            string          `json:"type"` // "file" | "folder"
	Size            int             `json:"size"`
	FileReference   string          `json:"fileReference,omitempty"`
	Text            string          `json:"text,omitempty"`
	CreatedDate     string          `json:"createdDate,omitempty"`
	LastUpdatedDate string          `json:"lastUpdatedDate,omitempty"`
	ParentFileID    string          `json:"parentFileId,omitempty"`
	FileFolderID    string          `json:"fileFolderId,omitempty"`
	Children        []*MetadataNode `json:"children,omitempty"`
}

// Metadata is the top-level document persisted to ~/.filesystem/metadata.json.
// It matches the JSON shape returned by the files-handler /files/metadata
// endpoint: a small set of identifiers plus the root of the file tree.
type Metadata struct {
	FileFolderID string        `json:"fileFolderId"`
	RootFileID   string        `json:"rootFileId"`
	MemberID     string        `json:"memberId,omitempty"`
	UserID       string        `json:"userId,omitempty"`
	GeneratedAt  string        `json:"generatedAt,omitempty"`
	Root         *MetadataNode `json:"root"`
}

// metadataFileName is the name of the cached metadata JSON file.
const metadataFileName = "metadata.json"

// MetadataPath returns the absolute path to the cached metadata JSON file.
func MetadataPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, metadataFileName), nil
}

// SaveMetadata persists raw JSON bytes returned by the /files/metadata
// endpoint to ~/.filesystem/metadata.json with 0600 permissions.
// The raw form is kept verbatim so future schema additions are preserved.
func SaveMetadata(raw []byte) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	// Pretty-print so the file is pleasant to inspect by hand.
	var pretty any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return fmt.Errorf("invalid metadata JSON: %w", err)
	}
	out, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}

	path := filepath.Join(dir, metadataFileName)
	if err := os.WriteFile(path, out, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// DeleteMetadata removes ~/.filesystem/metadata.json.
// Returns nil if the file does not exist.
func DeleteMetadata() error {
	path, err := MetadataPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	return nil
}

// LoadMetadata reads the cached metadata JSON and decodes it into a Metadata
// structure. If the file does not exist, (nil, nil) is returned so callers
// can degrade gracefully (e.g. shell completion).
func LoadMetadata() (*Metadata, error) {
	path, err := MetadataPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var md Metadata
	if err := json.Unmarshal(data, &md); err != nil {
		return nil, fmt.Errorf("parse metadata: %w", err)
	}
	return &md, nil
}

// ── Path navigation helpers ──────────────────────────────────────────────────

// ResolvePath walks the metadata tree and returns the node at the given
// slash-separated path. "/" (or an empty string) returns the root node.
// Matching is case-insensitive to match the CLI's existing behavior.
func (m *Metadata) ResolvePath(path string) *MetadataNode {
	if m == nil || m.Root == nil {
		return nil
	}
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return m.Root
	}

	node := m.Root
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." {
			continue
		}
		if node.Type != "folder" {
			return nil
		}
		var next *MetadataNode
		for _, child := range node.Children {
			if strings.EqualFold(child.Name, part) {
				next = child
				break
			}
		}
		if next == nil {
			return nil
		}
		node = next
	}
	return node
}

// CompletePath returns path completions for the given (possibly partial)
// argument. The returned strings are absolute paths (starting with "/")
// suitable for use as shell completion results.
//
// When includeFiles is false, only folders are returned (useful for commands
// like `move` where the destination must be a folder).
func (m *Metadata) CompletePath(arg string, includeFiles bool) []string {
	if m == nil || m.Root == nil {
		return nil
	}

	// Normalize leading slash. Users typically type "/Folder/sub"; treat
	// a missing leading slash as root-relative.
	hasLeading := strings.HasPrefix(arg, "/")
	clean := arg
	if !hasLeading {
		clean = "/" + clean
	}

	// Split into "directory" portion (parent folder to list) and "prefix"
	// portion (partial name to filter by). A trailing slash means the user
	// wants to see everything inside that folder.
	var parentPath, prefix string
	if strings.HasSuffix(clean, "/") {
		parentPath = clean
		prefix = ""
	} else {
		idx := strings.LastIndex(clean, "/")
		parentPath = clean[:idx+1] // include trailing "/"
		prefix = clean[idx+1:]
	}

	parent := m.ResolvePath(parentPath)
	if parent == nil || parent.Type != "folder" {
		return nil
	}

	// Preserve the user's original leading style: if they typed no "/",
	// don't add one in the completions either.
	parentDisplay := parentPath
	if !hasLeading {
		parentDisplay = strings.TrimPrefix(parentDisplay, "/")
	}

	lowerPrefix := strings.ToLower(prefix)
	var results []string
	for _, child := range parent.Children {
		if !includeFiles && child.Type != "folder" {
			continue
		}
		if lowerPrefix != "" && !strings.HasPrefix(strings.ToLower(child.Name), lowerPrefix) {
			continue
		}
		entry := parentDisplay + child.Name
		if child.Type == "folder" {
			entry += "/"
		}
		results = append(results, entry)
	}
	return results
}
