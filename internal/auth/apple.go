package auth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleTokenURL  = "https://appleid.apple.com/auth/token"
	appleRevokeURL = "https://appleid.apple.com/auth/revoke"
	appleAudience  = "https://appleid.apple.com"
)

type AppleClient struct {
	teamID   string
	keyID    string
	clientID string
	key      *ecdsa.PrivateKey
	http     *http.Client
}

func NewAppleClient(teamID, keyID, clientID, privateKeyPEM string) (*AppleClient, error) {
	keyPEM := strings.ReplaceAll(privateKeyPEM, `\n`, "\n")
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(keyPEM))
	if err != nil {
		return nil, fmt.Errorf("parsing apple private key: %w", err)
	}
	return &AppleClient{
		teamID:   teamID,
		keyID:    keyID,
		clientID: clientID,
		key:      key,
		http:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *AppleClient) RevokeWithAuthorizationCode(ctx context.Context, code string) error {
	secret, err := c.clientSecret()
	if err != nil {
		return fmt.Errorf("building apple client secret: %w", err)
	}

	body, err := c.postForm(ctx, appleTokenURL, url.Values{
		"client_id":     {c.clientID},
		"client_secret": {secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return err
	}

	var tokens struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokens); err != nil {
		return fmt.Errorf("decoding apple token response: %w", err)
	}

	token, hint := tokens.RefreshToken, "refresh_token"
	if token == "" {
		token, hint = tokens.AccessToken, "access_token"
	}
	if token == "" {
		return fmt.Errorf("apple token response contained no token")
	}

	_, err = c.postForm(ctx, appleRevokeURL, url.Values{
		"client_id":       {c.clientID},
		"client_secret":   {secret},
		"token":           {token},
		"token_type_hint": {hint},
	})
	return err
}

func (c *AppleClient) clientSecret() (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.RegisteredClaims{
		Issuer:    c.teamID,
		Subject:   c.clientID,
		Audience:  jwt.ClaimStrings{appleAudience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	})
	token.Header["kid"] = c.keyID
	return token.SignedString(c.key)
}

func (c *AppleClient) postForm(ctx context.Context, endpoint string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building apple request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling apple: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, fmt.Errorf("reading apple response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(body, &apiErr)
		return nil, fmt.Errorf("apple %s: status %d (%s)", endpoint, resp.StatusCode, apiErr.Error)
	}
	return body, nil
}
