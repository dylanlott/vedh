package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authContextKey struct{}

type AuthUser struct {
	ID       string
	Username string
}

type AuthClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func jwtSecret() ([]byte, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return []byte(secret), nil
}

func withAuth(ctx context.Context, user *AuthUser) context.Context {
	if user == nil {
		return ctx
	}
	return context.WithValue(ctx, authContextKey{}, user)
}

func authFromContext(ctx context.Context) (*AuthUser, bool) {
	user, ok := ctx.Value(authContextKey{}).(*AuthUser)
	return user, ok && user != nil
}

func parseBearer(headerVal string) string {
	if headerVal == "" {
		return ""
	}
	parts := strings.SplitN(headerVal, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parseAuthFromRequest(r *http.Request) (*AuthUser, error) {
	if r == nil {
		return nil, nil
	}
	token := parseBearer(r.Header.Get("Authorization"))
	if token == "" {
		return nil, nil
	}
	return parseAndValidateToken(token)
}

func parseAuthFromInitPayload(initPayload map[string]any) (*AuthUser, error) {
	if initPayload == nil {
		return nil, nil
	}
	raw, ok := initPayload["authorization"].(string)
	if !ok || raw == "" {
		return nil, nil
	}
	token := parseBearer(raw)
	if token == "" {
		return nil, errors.New("invalid authorization header")
	}
	return parseAndValidateToken(token)
}

func parseAndValidateToken(tokenString string) (*AuthUser, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}
	claims := &AuthClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Subject == "" {
		return nil, errors.New("invalid token subject")
	}
	return &AuthUser{
		ID:       claims.Subject,
		Username: claims.Username,
	}, nil
}

func newAuthToken(user *User, ttl time.Duration) (string, error) {
	if user == nil || user.ID == "" {
		return "", errors.New("invalid user for token")
	}
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := AuthClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
