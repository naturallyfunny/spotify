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

// TokenStore persists Spotify OAuth tokens on behalf of a user.
// Implement this interface to provide your own storage backend.
// GetRefreshToken must return ErrNotConnected when no token exists for userID.
type TokenStore interface {
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	// SaveRefreshToken stores (or replaces) the refresh token for userID.
	// It is called after a successful OAuth exchange.
	SaveRefreshToken(ctx context.Context, userID, refreshToken string) error
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

// AuthURL returns the Spotify Accounts authorization URL the user must visit
// to grant access. state is handed to Spotify and returned verbatim on the
// callback; consumers use it to correlate the callback with a user and to
// guard against CSRF. The redirect URI and scopes are taken from the
// Authenticator supplied to New.
func (c *Client) AuthURL(state string) string {
	return c.auth.AuthURL(state)
}

// Exchange completes the OAuth flow by trading the authorization code from the
// Spotify callback for tokens, returning the refresh token to persist via
// TokenStore.SaveRefreshToken. The code is single-use; a second Exchange with
// the same code is rejected by Spotify.
func (c *Client) Exchange(ctx context.Context, code string) (string, error) {
	token, err := c.auth.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	if token.RefreshToken == "" {
		return "", errors.New("spotify: exchange returned no refresh token")
	}
	return token.RefreshToken, nil
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
