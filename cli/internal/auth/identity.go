package auth

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
)

const (
	// identityPoolID is the Cognito Identity Pool ID from amplify_outputs.json.
	identityPoolID = "us-east-1:144cc6cc-8376-45ff-a651-7c3644eca806"
	// userPoolID is the Cognito User Pool ID from amplify_outputs.json.
	userPoolID = "us-east-1_gI0gO1dL0"
)

// AWSCredentials holds temporary AWS credentials obtained from the Cognito
// Identity Pool, plus the resolved identity ID.
type AWSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	IdentityID      string // Cognito Identity Pool identity ID (session.identityId)
}

// GetAWSCredentials exchanges a Cognito ID token for temporary AWS credentials
// via the Identity Pool.  This mirrors the Amplify fetchAuthSession() flow used
// by the React app to get credentials for S3 uploads.
func GetAWSCredentials(ctx context.Context, idToken string) (*AWSCredentials, error) {
	// Use anonymous base credentials – the Identity Pool exchange itself is the
	// authentication step; no IAM signing is required for the Cognito Identity
	// API calls.
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(awsRegion),
		awsconfig.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return nil, fmt.Errorf("configure AWS SDK: %w", err)
	}

	idClient := cognitoidentity.NewFromConfig(cfg)

	// Logins map: associates the ID token with the Cognito User Pool.
	loginsKey := fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s", awsRegion, userPoolID)
	logins := map[string]string{loginsKey: idToken}

	// Step 1 – resolve the Cognito Identity ID for this user.
	getIDOut, err := idClient.GetId(ctx, &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String(identityPoolID),
		Logins:         logins,
	})
	if err != nil {
		return nil, fmt.Errorf("get identity ID: %w", err)
	}

	identityID := aws.ToString(getIDOut.IdentityId)

	// Step 2 – exchange the identity ID for temporary AWS credentials.
	getCredsOut, err := idClient.GetCredentialsForIdentity(ctx, &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: aws.String(identityID),
		Logins:     logins,
	})
	if err != nil {
		return nil, fmt.Errorf("get credentials for identity: %w", err)
	}

	c := getCredsOut.Credentials
	return &AWSCredentials{
		AccessKeyID:     aws.ToString(c.AccessKeyId),
		SecretAccessKey: aws.ToString(c.SecretKey),
		SessionToken:    aws.ToString(c.SessionToken),
		IdentityID:      identityID,
	}, nil
}
