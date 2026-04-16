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

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with filesystem.io",
	Long: `Log in to filesystem.io using your email and password.

Credentials are stored securely in ~/.filesystem/credentials.json.

Note: USER_PASSWORD_AUTH must be enabled in the Cognito User Pool client
settings for this command to work. This can be enabled in the AWS Console
under Cognito > User Pools > App clients > Edit.`,
	RunE: runLogin,
}

func runLogin(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	password := string(passwordBytes)
	if password == "" {
		return fmt.Errorf("password cannot be empty")
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
