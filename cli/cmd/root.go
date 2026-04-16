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
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Wrap commands that require authentication so that a 401 response
	// automatically triggers the login flow and retries the command.
	listCmd.RunE = withAutoLogin(runList)
	downloadCmd.RunE = withAutoLogin(runDownload)
	renameCmd.RunE = withAutoLogin(runRename)
	createCmd.RunE = withAutoLogin(runCreate)
	moveCmd.RunE = withAutoLogin(runMove)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(renameCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(moveCmd)
}
