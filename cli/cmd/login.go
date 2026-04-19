package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Environment variables that, when both are set, allow the login flow to
// run non-interactively. This is useful for scripts, CI pipelines, and
// automation where prompting for input is not possible.
const (
	envUsername = "FS_USERNAME"
	envPassword = "FS_PASSWORD"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with filesystem.io",
	Long: `Log in to filesystem.io using your email and password.

Credentials are stored securely in ~/.filesystem/credentials.json.

For non-interactive use (e.g. scripts or CI), set the FS_USERNAME and
FS_PASSWORD environment variables and the login command will use those
values instead of prompting.

Note: USER_PASSWORD_AUTH must be enabled in the Cognito User Pool client
settings for this command to work. This can be enabled in the AWS Console
under Cognito > User Pools > App clients > Edit.`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	email, password, err := readCredentials()
	if err != nil {
		return err
	}

	fmt.Println("Authenticating...")

	tokens, err := auth.Login(context.Background(), email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	creds := &config.Credentials{
		AccessToken:  tokens.AccessToken,
		IDToken:      tokens.IDToken,
		RefreshToken: tokens.RefreshToken,
	}

	if err := config.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println("✓ Logged in successfully!")
	return nil
}

// readCredentials resolves the email and password used for login.
//
// If both FS_USERNAME and FS_PASSWORD environment variables are set, they
// are used directly and no prompts are shown. Otherwise, the user is
// prompted interactively (with individual values still overridable via
// their matching environment variable).
func readCredentials() (string, string, error) {
	envEmail := strings.TrimSpace(os.Getenv(envUsername))
	envPass := os.Getenv(envPassword)

	// Fully non-interactive path: both env vars set.
	if envEmail != "" && envPass != "" {
		//fmt.Printf("Using credentials from %s / %s environment variables.\n", envUsername, envPassword)
		return envEmail, envPass, nil
	}

	reader := bufio.NewReader(os.Stdin)

	email := envEmail
	if email == "" {
		fmt.Print("Email: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", "", fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(input)
	}
	if email == "" {
		return "", "", fmt.Errorf("email cannot be empty")
	}

	password := envPass
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", "", fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println()
		password = string(passwordBytes)
	}
	if password == "" {
		return "", "", fmt.Errorf("password cannot be empty")
	}

	return email, password, nil
}
