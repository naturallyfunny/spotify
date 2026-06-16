package spotify

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// TokenStore retrieves Spotify OAuth tokens on behalf of a user.
// Implement this interface to provide your own storage backend.
type TokenStore interface {
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}

// Client provides access to Spotify on behalf of a user.
// It delegates token storage to a TokenStore and API calls to the zmb3 client.
type Client struct {
	tokenStore TokenStore
	auth       *spotifyauth.Authenticator
}

func New(tokenStore TokenStore, auth *spotifyauth.Authenticator) *Client {
	return &Client{
		tokenStore: tokenStore,
		auth:       auth,
	}
}

// Track represents a Spotify track.
type Track struct {
	ID      string
	Name    string
	Artists []string
	URI     string
	URL     string
}

// Playlist represents a Spotify playlist.
type Playlist struct {
	ID          string
	Name        string
	Description string
	Total       int
	URL         string
}

// Device represents an active Spotify playback device.
type Device struct {
	ID       string
	Name     string
	Type     string
	IsActive bool
	Volume   int
}

// clientFor creates an authenticated zmb3 client for the given user.
func (c *Client) clientFor(ctx context.Context, userID string) (*spotify.Client, error) {
	refreshToken, err := c.tokenStore.GetRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get token for user %s: %w", userID, err)
	}
	token := &oauth2.Token{RefreshToken: refreshToken}
	httpClient := c.auth.Client(ctx, token)
	return spotify.New(httpClient), nil
}
