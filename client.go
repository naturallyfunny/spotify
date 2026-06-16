package spotify

import (
	"context"
	"errors"
	"fmt"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// ErrNotConnected indicates the user has no stored Spotify credentials,
// i.e. they have not completed the OAuth connect flow. TokenStore
// implementations should return this from GetRefreshToken when no token
// exists, so consumers can route the user into the login flow.
var ErrNotConnected = errors.New("spotify: user not connected")

// TokenStore retrieves Spotify OAuth tokens on behalf of a user.
// Implement this interface to provide your own storage backend.
// GetRefreshToken must return ErrNotConnected when no token exists for userID.
type TokenStore interface {
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}

// Client provides access to Spotify on behalf of a user.
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

func (c *Client) clientFor(ctx context.Context, userID string) (*spotify.Client, error) {
	refreshToken, err := c.tokenStore.GetRefreshToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get token for user %s: %w", userID, err)
	}
	token := &oauth2.Token{RefreshToken: refreshToken}
	httpClient := c.auth.Client(ctx, token)
	return spotify.New(httpClient), nil
}
