package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)


type AdminClient struct {
	baseURL   string
	secretKey string
	http      *http.Client
}

func NewAdminClient(baseURL, secretKey string) *AdminClient {
	return &AdminClient{
		baseURL:   strings.TrimRight(baseURL, "/"),
		secretKey: secretKey,
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *AdminClient) DeleteUser(ctx context.Context, userID string) error {
	endpoint := fmt.Sprintf("%s/auth/v1/admin/users/%s", c.baseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("building delete user request: %w", err)
	}

	req.Header.Set("apikey", c.secretKey)

	if !strings.HasPrefix(c.secretKey, "sb_secret_") {
		req.Header.Set("Authorization", "Bearer "+c.secretKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("calling supabase admin api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil
	}
	return fmt.Errorf("supabase admin delete user: unexpected status %d", resp.StatusCode)
}
