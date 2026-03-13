package satusehat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model/satusehat"
	"github.com/faisalhardin/medilink/internal/entity/repo/cache"
)

// Client handles authenticated HTTP requests to Satu Sehat FHIR API
type Client struct {
	authClient *AuthClient
	httpClient *http.Client
	baseURL    string
	orgID      string
}

// NewClient creates a new Satu Sehat FHIR API client
func NewClient(cfg *config.Config, cache cache.Caching) *Client {
	return &Client{
		authClient: NewAuthClient(cfg, cache),
		httpClient: &http.Client{Timeout: 60 * time.Second},
		baseURL:    cfg.SatuSehatConfig.BaseURL,
		orgID:      cfg.SatuSehatConfig.OrganizationID,
	}
}

// DoRequest performs an authenticated HTTP request to Satu Sehat FHIR API
// Automatically injects OAuth2 access token and handles 401 errors with retry
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Build request URL
	requestURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get access token
	token, err := c.authClient.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Handle 401 Unauthorized - token might be expired, retry once with new token
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		// Invalidate cached token
		c.authClient.InvalidateToken(ctx)

		// Get new token
		token, err = c.authClient.GetAccessToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get new access token: %w", err)
		}

		// Retry request with new token
		if body != nil {
			bodyBytes, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err = http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute retry request: %w", err)
		}
	}

	return resp, nil
}

// Get performs GET request to retrieve a specific resource by ID
func (c *Client) Get(ctx context.Context, resourceType, id string, result interface{}) error {
	path := fmt.Sprintf("/%s/%s", resourceType, id)
	resp, err := c.DoRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// Post performs POST request to create a new resource
func (c *Client) Post(ctx context.Context, resourceType string, resource interface{}, result interface{}) error {
	path := fmt.Sprintf("/%s", resourceType)
	resp, err := c.DoRequest(ctx, "POST", path, resource)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Put performs PUT request to update an existing resource
func (c *Client) Put(ctx context.Context, resourceType, id string, resource interface{}, result interface{}) error {
	path := fmt.Sprintf("/%s/%s", resourceType, id)
	resp, err := c.DoRequest(ctx, "PUT", path, resource)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Delete performs DELETE request to remove a resource
func (c *Client) Delete(ctx context.Context, resourceType, id string) error {
	path := fmt.Sprintf("/%s/%s", resourceType, id)
	resp, err := c.DoRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	return nil
}

// Search performs search query on a resource type
func (c *Client) Search(ctx context.Context, resourceType string, params url.Values, result *satusehat.Bundle) error {
	path := fmt.Sprintf("/%s", resourceType)
	if len(params) > 0 {
		path = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	resp, err := c.DoRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode search response: %w", err)
	}

	return nil
}

// PostBundle submits a FHIR Bundle (batch or transaction)
func (c *Client) PostBundle(ctx context.Context, bundle *satusehat.Bundle) (*satusehat.Bundle, error) {
	resp, err := c.DoRequest(ctx, "POST", "", bundle)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var resultBundle satusehat.Bundle
	if err := json.NewDecoder(resp.Body).Decode(&resultBundle); err != nil {
		return nil, fmt.Errorf("failed to decode bundle response: %w", err)
	}

	return &resultBundle, nil
}

// Validate validates a resource without persisting it
func (c *Client) Validate(ctx context.Context, resourceType string, resource interface{}) error {
	path := fmt.Sprintf("/%s/$validate", resourceType)
	resp, err := c.DoRequest(ctx, "POST", path, resource)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.parseError(resp)
	}

	var outcome satusehat.OperationOutcome
	if err := json.NewDecoder(resp.Body).Decode(&outcome); err != nil {
		return fmt.Errorf("failed to decode validation response: %w", err)
	}

	// Check for errors in validation outcome
	for _, issue := range outcome.Issue {
		if issue.Severity == "error" || issue.Severity == "fatal" {
			return fmt.Errorf("validation failed: %s - %s", issue.Code, issue.Diagnostics)
		}
	}

	return nil
}

// parseError parses FHIR OperationOutcome error responses
func (c *Client) parseError(resp *http.Response) error {
	var outcome satusehat.OperationOutcome
	if err := json.NewDecoder(resp.Body).Decode(&outcome); err != nil {
		return fmt.Errorf("request failed with status %d: unable to parse error response", resp.StatusCode)
	}

	if len(outcome.Issue) > 0 {
		issue := outcome.Issue[0]
		return fmt.Errorf("FHIR error (%s): %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
	}

	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}

// GetOrganizationID returns the configured organization ID
func (c *Client) GetOrganizationID() string {
	return c.orgID
}
