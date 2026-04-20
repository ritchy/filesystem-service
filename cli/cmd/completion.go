package cmd

import (
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
)

// completePath returns a cobra completion function that suggests remote
// filesystem paths from the cached metadata (~/.filesystem/metadata.json).
//
// Behavior:
//   - When no metadata cache exists, no completions are returned (so shell
//     defaults such as file-name completion can kick in).
//   - includeFiles controls whether files are suggested in addition to
//     folders (set to false for commands whose argument must be a folder).
//
// The NoSpace directive keeps the prompt on the same "word" after completing
// a folder (which ends with "/") so the user can immediately tab again to
// descend into it.
func completePath(includeFiles bool) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		md, err := config.LoadMetadata()
		if err != nil || md == nil {
			// Gracefully fall back to default (e.g. filename) completion.
			return nil, cobra.ShellCompDirectiveDefault
		}

		matches := md.CompletePath(toComplete, includeFiles)

		// Suppress automatic space insertion so users can keep tabbing into
		// folders (completions ending in "/"). Also turn off file completion
		// for the path argument since it refers to a remote path.
		return matches, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}
}

// completeRemotePathPos returns a completion function that only activates
// for the given 0-based argument position. Used by commands like `move`
// (where both positional args are remote paths) and `upload` (only the
// second positional arg is a remote path; the first is a local file).
func completeRemotePathPos(pos int, includeFiles bool) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != pos {
			// Let the default (file) completion handle it.
			return nil, cobra.ShellCompDirectiveDefault
		}
		return completePath(includeFiles)(cmd, args, toComplete)
	}
}
