package spotify

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// ErrNotConnected indicates the user has no stored Spotify credentials,
// i.e. they have not completed the OAuth connect flow. TokenStore
// implementations should return this from GetRefreshToken when no token
// exists, so consumers can route the user into the login flow.
var ErrNotConnected = errors.New("spotify: user not connected")

// ErrMissingScopes indicates the user completed the OAuth flow but did not
// grant every scope the library needs (see RequiredScopes). Detected at
// Exchange time so consumers can fail at connect, not months later when a
// playback call returns 403. Inspect the granted/missing sets via a
// *ScopeError obtained with errors.As.
var ErrMissingScopes = errors.New("spotify: missing required scopes")

// RequiredScopes are the OAuth scopes every capability in this library needs.
// Pass them to the Authenticator so the consent screen requests the right
// permissions:
//
//	auth := spotifyauth.New(
//	    spotifyauth.WithClientID(id),
//	    spotifyauth.WithClientSecret(secret),
//	    spotifyauth.WithRedirectURL(redirectURL),
//	    spotifyauth.WithScopes(spotify.RequiredScopes...),
//	)
//
// Mapping of scope to capability:
//   - user-modify-playback-state: Play, Pause, Resume, SetVolume
//   - user-read-playback-state:   Devices
//   - playlist-read-private:      UserPlaylists
//
// Search (SearchTracks, SearchPlaylists, PlaylistTracks) needs no scope.
var RequiredScopes = []string{
	spotifyauth.ScopeUserModifyPlaybackState,
	spotifyauth.ScopeUserReadPlaybackState,
	spotifyauth.ScopePlaylistReadPrivate,
}

// ScopeError reports which scopes were granted versus which are missing after
// an OAuth exchange. It wraps ErrMissingScopes; match it with errors.As.
type ScopeError struct {
	Granted []string
	Missing []string
}

func (e *ScopeError) Error() string {
	return fmt.Sprintf("spotify: missing required scopes %v (granted %v) — reconnect granting full permissions", e.Missing, e.Granted)
}

func (e *ScopeError) Unwrap() error { return ErrMissingScopes }

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
	granted := grantedScopes(token)
	if missing := missingScopes(RequiredScopes, granted); len(missing) > 0 {
		return "", &ScopeError{Granted: granted, Missing: missing}
	}
	return token.RefreshToken, nil
}

// grantedScopes reads the scopes Spotify actually granted from the token
// exchange response. Spotify returns them as a space-separated "scope" field,
// which oauth2 surfaces via Token.Extra. Reflects what the user approved, not
// merely what was requested.
func grantedScopes(token *oauth2.Token) []string {
	raw, _ := token.Extra("scope").(string)
	return strings.Fields(raw)
}

// missingScopes returns the elements of required not present in granted.
func missingScopes(required, granted []string) []string {
	have := make(map[string]struct{}, len(granted))
	for _, s := range granted {
		have[s] = struct{}{}
	}
	var missing []string
	for _, s := range required {
		if _, ok := have[s]; !ok {
			missing = append(missing, s)
		}
	}
	return missing
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
