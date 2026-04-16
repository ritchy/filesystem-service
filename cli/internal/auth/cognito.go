// Package auth handles Cognito authentication for the filesystem.io CLI.
//
// Prerequisites:
//
//	USER_PASSWORD_AUTH must be enabled on the Cognito User Pool App Client.
//	To enable it: AWS Console → Cognito → User Pools → App clients → Edit
//	→ Authentication flows → check "ALLOW_USER_PASSWORD_AUTH".
package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

const (
	// clientID is the Cognito User Pool App Client ID from amplify_outputs.json.
	clientID = "<redacted>"
	// awsRegion is the AWS region where the User Pool is hosted.
	awsRegion = "us-east-1"
)

// Tokens holds the three tokens returned by a successful Cognito authentication.
type Tokens struct {
	AccessToken  string
	IDToken      string
	RefreshToken string
}

// Login authenticates the given email/password against the Cognito User Pool
// and returns the resulting JWT tokens.
func Login(ctx context.Context, email, password string) (*Tokens, error) {
	// Use anonymous credentials – InitiateAuth with USER_PASSWORD_AUTH does not
	// require IAM signing; the username/password in the request body IS the auth.
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure AWS SDK: %w", err)
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	out, err := client.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitotypes.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(clientID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Handle any follow-up challenges
	if out.ChallengeName != "" {
		switch out.ChallengeName {
		case cognitotypes.ChallengeNameTypeNewPasswordRequired:
			return nil, fmt.Errorf(
				"a new password is required – please sign in via the web app first to set one",
			)
		default:
			return nil, fmt.Errorf("unexpected challenge: %s", out.ChallengeName)
		}
	}

	if out.AuthenticationResult == nil {
		return nil, fmt.Errorf("no authentication result returned from Cognito")
	}

	return &Tokens{
		AccessToken:  aws.ToString(out.AuthenticationResult.AccessToken),
		IDToken:      aws.ToString(out.AuthenticationResult.IdToken),
		RefreshToken: aws.ToString(out.AuthenticationResult.RefreshToken),
	}, nil
}
