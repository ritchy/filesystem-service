// Package api provides a GraphQL client for the filesystem.io AppSync API.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// graphQLEndpoint is the AppSync GraphQL URL from amplify_outputs.json.
const graphQLEndpoint = "https://zkfwgdtvibdpvbn6td4dcurghq.appsync-api.us-east-1.amazonaws.com/graphql"

// filesAPIEndpoint is the REST API Gateway URL for the files handler.
const filesAPIEndpoint = "https://i7w0p5qieb.execute-api.us-east-1.amazonaws.com/dev/"

// Client is an authenticated GraphQL client for the filesystem.io AppSync API.
type Client struct {
	idToken    string
	httpClient *http.Client
}

// NewClient creates a Client that authenticates using the provided Cognito ID token.
func NewClient(idToken string) *Client {
	return &Client{
		idToken:    idToken,
		httpClient: &http.Client{},
	}
}

// ── internal types ────────────────────────────────────────────────────────────

type gqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors,omitempty"`
}

type gqlError struct {
	Message string `json:"message"`
}

// execute sends a GraphQL request and unmarshals the response "data" field
// into result (which must be a pointer).
func (c *Client) execute(
	ctx context.Context,
	query string,
	variables map[string]interface{},
	result interface{},
) error {
	body, err := json.Marshal(gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, graphQLEndpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// AppSync Cognito User Pools auth: pass the raw ID token as Authorization.
	req.Header.Set("Authorization", c.idToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp gqlResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("parse GraphQL response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if result != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}

// GetFileInfo fetches metadata for a file or folder from the /info/:id REST
// endpoint.  The response mirrors the React app's fetchFileInfo() call.
func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*FileInfo, error) {
	url := fmt.Sprintf("%sinfo/%s", filesAPIEndpoint, fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create info request: %w", err)
	}
	req.Header.Set("Authorization", c.idToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var info FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode info response: %w", err)
	}

	return &info, nil
}

// DownloadDirect fetches file data from the /direct REST endpoint.
// The endpoint returns either a 302 redirect to a pre-signed S3 URL (binary
// files) or a 200 text/plain body (text files stored inline).  Go's default
// http.Client follows redirects automatically; the Authorization header is
// stripped before following cross-origin redirects so the S3 pre-signed URL
// is not polluted by an extra auth header.
func (c *Client) DownloadDirect(ctx context.Context, fileID string) ([]byte, error) {
	url := fmt.Sprintf("%sdirect?id=%s", filesAPIEndpoint, fileID)

	// Build a one-shot HTTP client that drops the Authorization header when
	// following a redirect to a different host (i.e., to S3).
	downloadClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && req.URL.Host != via[0].URL.Host {
				req.Header.Del("Authorization")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}
	req.Header.Set("Authorization", c.idToken)

	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("file not found on server")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download response: %w", err)
	}

	return data, nil
}
