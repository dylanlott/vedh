package server

import (
	"context"
	"errors"
)

var publicQueries = map[string]struct{}{
	"card":      {},
	"cards":     {},
	"search":    {},
	"searchAll": {},
}

func isPublicQuery(name string) bool {
	_, ok := publicQueries[name]
	return ok
}

func requireAuth(ctx context.Context) (*AuthUser, error) {
	if user, ok := authFromContext(ctx); ok {
		return user, nil
	}
	return nil, errors.New("authentication required")
}

func requireMatchingUser(ctx context.Context, userID string, username string) (*AuthUser, error) {
	user, err := requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if userID != "" && user.ID != userID {
		return nil, errors.New("forbidden: user mismatch")
	}
	if username != "" && user.Username != "" && user.Username != username {
		return nil, errors.New("forbidden: username mismatch")
	}
	return user, nil
}

func isUserInGame(game *Game, user *AuthUser) bool {
	if game == nil || user == nil {
		return false
	}
	for _, p := range game.Players {
		if p == nil {
			continue
		}
		if user.ID != "" && p.ID == user.ID {
			return true
		}
		if user.Username != "" && p.Username == user.Username {
			return true
		}
	}
	return false
}
