package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type AMREntry struct {
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
}

type Claims struct {
	jwt.RegisteredClaims
	Email string     `json:"email"`
	AMR   []AMREntry `json:"amr"`
}

func (c *Claims) LastSignInAt() time.Time {
	var latest int64
	for _, entry := range c.AMR {
		if entry.Timestamp > latest {
			latest = entry.Timestamp
		}
	}
	if latest == 0 {
		return time.Time{}
	}
	return time.Unix(latest, 0)
}

type Verifier struct {
	keyfunc keyfunc.Keyfunc
}

func NewVerifier(ctx context.Context, jwksURL string) (*Verifier, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("fetching jwks: %w", err)
	}
	return &Verifier{keyfunc: kf}, nil
}

func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, v.keyfunc.Keyfunc)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if aud, _ := claims.GetAudience(); len(aud) == 0 || aud[0] != "authenticated" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
