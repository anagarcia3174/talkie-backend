package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
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