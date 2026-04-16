package cmd

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"

	"filesystem-cli/internal/api"
	"filesystem-cli/internal/auth"
	"filesystem-cli/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

const (
	// s3Bucket is the Amplify storage bucket name from amplify_outputs.json.
	s3Bucket = "amplify-d3mmq1faysmf0w-ma-filesystemstoragebucket1-mxyzl4voiav3"
	// s3Region is the AWS region of the S3 bucket.
	s3Region = "us-east-1"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <local-file> <remote-folder-path>",
	Short: "Upload a local file to your filesystem",
	Long: `Upload a local file to the specified folder in your filesystem.

The file is stored in S3 and a File entry is created in the database,
mirroring the behaviour of the React app's uploadFile() function.

Examples:
  fs upload report.pdf /documents          # upload to /documents
  fs upload notes.txt  /                   # upload to root folder
  fs upload photo.png  /photos/2024        # upload to a nested folder`,
	Args: cobra.ExactArgs(2),
	RunE: runUpload,
}

func runUpload(cmd *cobra.Command, args []string) error {
	localPath := args[0]
	remotePath := args[1]

	// ── Read the local file ────────────────────────────────────────────────────
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %w", localPath, err)
	}

	fileName := filepath.Base(localPath)

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

	// ── Resolve user root ─────────────────────────────────────────────────────
	member, err := apiClient.GetMemberByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve account info: %w", err)
	}

	if member.FileFolder.ID == "" {
		return fmt.Errorf("no file system found for your account")
	}

	rootFileID := member.FileFolder.RootFileID
	fileFolderID := member.FileFolder.ID

	// ── Resolve destination folder ────────────────────────────────────────────
	parentFolderID, err := apiClient.NavigatePath(ctx, rootFileID, remotePath)
	if err != nil {
		return fmt.Errorf("destination folder not found: %s (%v)", remotePath, err)
	}

	// ── Get AWS credentials via Cognito Identity Pool ─────────────────────────
	// Mirrors fetchAuthSession() → session.identityId || session.userSub
	fmt.Printf("  Authenticating with AWS…\n")
	awsCreds, err := auth.GetAWSCredentials(ctx, creds.IDToken)
	if err != nil {
		return fmt.Errorf("failed to obtain AWS credentials: %w", err)
	}

	// Use the Identity Pool identity ID as the user identifier in the S3 path
	// (equivalent to session.identityId in the React app).
	s3UserID := awsCreds.IdentityID
	if s3UserID == "" {
		s3UserID = userID // fall back to Cognito sub (session.userSub)
	}

	// ── Build S3 path ─────────────────────────────────────────────────────────
	// Format: "files/${userId}/${Date.now()}_${fileName}"
	s3Key := fmt.Sprintf("files/%s/%d_%s", s3UserID, time.Now().UnixMilli(), fileName)

	// ── Upload to S3 ──────────────────────────────────────────────────────────
	fmt.Printf("  Uploading %s (%s) → s3://%s/%s\n", fileName, formatSize(len(data)), s3Bucket, s3Key)

	if err := putS3Object(ctx, awsCreds, s3Key, fileName, data); err != nil {
		return fmt.Errorf("S3 upload failed: %w", err)
	}

	// ── Create File entry in the database ─────────────────────────────────────
	created, err := apiClient.CreateFile(ctx, parentFolderID, fileFolderID, fileName, s3Key, len(data))
	if err != nil {
		return fmt.Errorf("failed to create file record: %w", err)
	}

	fmt.Printf("\n  ✓ Uploaded %q → %s  (id: %s)\n\n", fileName, remotePath, created.ID)
	return nil
}

// putS3Object uploads data to S3 using temporary credentials from the Identity
// Pool, replicating the uploadData() call in the React app.
func putS3Object(ctx context.Context, awsCreds *auth.AWSCredentials, key, fileName string, data []byte) error {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(s3Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				awsCreds.AccessKeyID,
				awsCreds.SecretAccessKey,
				awsCreds.SessionToken,
			),
		),
	)
	if err != nil {
		return fmt.Errorf("configure S3 client: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// Detect MIME type from the file extension; default to octet-stream.
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("PutObject: %w", err)
	}

	return nil
}
