package satusehat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model/satusehat"
	"github.com/faisalhardin/medilink/internal/entity/repo/cache"
)

// AuthClient handles OAuth2 authentication with Satu Sehat
type AuthClient struct {
	config       *config.Config
	cache        cache.Caching
	httpClient   *http.Client
	baseURL      string
	oauth2URL    string
	clientID     string
	clientSecret string
	mu           sync.Mutex // Prevents thundering herd on token refresh
}

// NewAuthClient creates a new Satu Sehat authentication client
func NewAuthClient(cfg *config.Config, cache cache.Caching) *AuthClient {
	return &AuthClient{
		config:       cfg,
		cache:        cache,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      cfg.SatuSehatConfig.BaseURL,
		oauth2URL:    cfg.SatuSehatConfig.OAuth2URL,
		clientID:     cfg.Vault.SatuSehatAuth.ClientID,
		clientSecret: cfg.Vault.SatuSehatAuth.ClientSecret,
	}
}

// GetAccessToken retrieves a valid access token (from cache or by requesting new one)
// This method is thread-safe and prevents multiple simultaneous token requests
func (c *AuthClient) GetAccessToken(ctx context.Context) (string, error) {
	// Check if integration is enabled
	if !c.config.SatuSehatConfig.Enabled {
		return "", fmt.Errorf("satu sehat integration is not enabled")
	}

	// Try to get cached token
	cacheKey := c.getCacheKey()
	cachedToken, err := c.cache.Get(cacheKey)
	if err == nil && cachedToken != "" {
		return cachedToken, nil
	}

	// Use mutex to prevent thundering herd
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache again after acquiring lock (another goroutine might have refreshed)
	cachedToken, err = c.cache.Get(cacheKey)
	if err == nil && cachedToken != "" {
		return cachedToken, nil
	}

	// Request new token
	tokenResp, err := c.RequestNewToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to request new token: %w", err)
	}

	// Cache the token with TTL (expires_in - 60 seconds buffer)
	ttl := tokenResp.ExpiresIn - 60
	if ttl < 0 {
		ttl = 0
	}
	_, err = c.cache.SetWithExpire(cacheKey, tokenResp.AccessToken, ttl)
	if err != nil {
		// Log error but don't fail - we still have the token
		fmt.Printf("Warning: failed to cache token: %v\n", err)
	}

	return tokenResp.AccessToken, nil
}

// RequestNewToken requests a new access token from Satu Sehat OAuth2 endpoint
func (c *AuthClient) RequestNewToken(ctx context.Context) (*satusehat.TokenResponse, error) {
	// Prepare request body
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	// Build request
	tokenURL := fmt.Sprintf("%s/accesstoken?grant_type=client_credentials", c.oauth2URL)
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseAuthError(resp)
	}

	// Parse response
	var tokenResp satusehat.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// ValidateToken checks if a token is valid (not expired)
func (c *AuthClient) ValidateToken(token *satusehat.TokenResponse) bool {
	if token == nil || token.AccessToken == "" {
		return false
	}

	// If IssuedAt is provided, calculate expiry time
	if token.IssuedAt != "" {
		issuedAt, err := time.Parse(time.RFC3339, token.IssuedAt)
		if err == nil {
			expiryTime := issuedAt.Add(time.Duration(token.ExpiresIn) * time.Second)
			// Add 60 second buffer before expiry
			return time.Now().Before(expiryTime.Add(-60 * time.Second))
		}
	}

	// If we can't determine expiry, assume invalid
	return false
}

// InvalidateToken removes the cached token, forcing a refresh on next request
func (c *AuthClient) InvalidateToken(ctx context.Context) error {
	cacheKey := c.getCacheKey()
	_, err := c.cache.Del(cacheKey)
	return err
}

// getCacheKey returns the cache key for the access token
func (c *AuthClient) getCacheKey() string {
	return fmt.Sprintf("satusehat:access_token:%s", c.clientID)
}

// parseAuthError parses authentication error responses
func (c *AuthClient) parseAuthError(resp *http.Response) error {
	var errResp satusehat.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("authentication failed with status %d: unable to parse error response", resp.StatusCode)
	}

	if len(errResp.Issue) > 0 {
		issue := errResp.Issue[0]
		return fmt.Errorf("authentication failed (%s): %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
	}

	return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
}
