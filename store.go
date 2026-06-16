package spotify

import "context"

// TokenStore retrieves Spotify OAuth tokens on behalf of a user.
// Implement this interface to provide your own storage backend.
type TokenStore interface {
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}
