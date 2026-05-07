package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fs",
	Short: "filesystem.io CLI tool",
	Long:  `A command-line interface for interacting with filesystem.io.`,
	// SilenceUsage prevents cobra from printing the usage/help text on
	// errors – we handle error presentation ourselves (especially for JSON).
	SilenceUsage: true,
	// SilenceErrors prevents cobra from printing errors to stderr so that
	// in JSON mode we can emit a clean JSON object instead.
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if JSONOutputEnabled {
			printJSONError("", err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func init() {
	// Global --json flag available to every sub-command.
	rootCmd.PersistentFlags().BoolVar(&JSONOutputEnabled, "json", false, "output results as JSON")

	// Wrap commands that require authentication so that a 401 response
	// automatically triggers the login flow and retries the command.
	listCmd.RunE = withAutoLogin(runList)
	downloadCmd.RunE = withAutoLogin(runDownload)
	renameCmd.RunE = withAutoLogin(runRename)
	createCmd.RunE = withAutoLogin(runCreate)
	moveCmd.RunE = withAutoLogin(runMove)
	uploadCmd.RunE = withAutoLogin(runUpload)
	shareCmd.RunE = withAutoLogin(runShare)
	deleteCmd.RunE = withAutoLogin(runDelete)
	fetchCmd.RunE = withAutoLogin(runFetch)
	refreshCmd.RunE = withAutoLogin(runRefresh)

	// Wire up tab-completion for commands that accept remote filesystem
	// paths. These read ~/.filesystem/metadata.json (populated by
	// `fs refresh`) and return matching folders/files.

	// list [path] – files and folders are both valid.
	listCmd.ValidArgsFunction = completePath(true)

	// download <path> – must be a file, but showing folders helps the user
	// descend into sub-directories during tab-completion.
	downloadCmd.ValidArgsFunction = completePath(true)

	// rename <source-path> <new-path> – both args refer to remote paths.
	// The destination is a new name, so no completions are offered for it.
	renameCmd.ValidArgsFunction = completeRemotePathPos(0, true)

	// create <path> – completions suggest the parent folder; the final
	// segment is the new name being created.
	createCmd.ValidArgsFunction = completePath(false)

	// move <source-path> <destination-folder-path> – source is anything,
	// destination must be a folder.
	moveCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completePath(true)(cmd, args, toComplete)
		case 1:
			return completePath(false)(cmd, args, toComplete)
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// upload <local-file> <remote-folder-path> – first arg is a local file
	// (use default file completion); second arg is a remote folder.
	uploadCmd.ValidArgsFunction = completeRemotePathPos(1, false)

	// share <file-path> – files and folders shown so user can descend.
	shareCmd.ValidArgsFunction = completePath(true)

	// delete <path> – files and folders are both valid.
	deleteCmd.ValidArgsFunction = completePath(true)

	// fetch <url> [remote-folder-path] – first arg is a URL (no completion);
	// second arg is an optional remote folder.
	fetchCmd.ValidArgsFunction = completeRemotePathPos(1, false)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(refreshCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(moveCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(shareCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(fetchCmd)
}
