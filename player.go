package spotify

import (
	"context"
	"errors"
	"strings"

	"github.com/zmb3/spotify/v2"
)

// These errors arise only from playback control, so they live with it. They are
// produced by sentinelFor (see client.go) and matched with errors.Is; the
// original Spotify error remains in the chain.
var (
	// ErrNoActiveDevice means Spotify has no device to act on: the user must
	// open Spotify somewhere before playback commands will land. Spotify
	// reports this as HTTP 404 with reason NO_ACTIVE_DEVICE.
	ErrNoActiveDevice = errors.New("spotify: no active device")

	// ErrPremiumRequired means the action needs a Spotify Premium account.
	// All playback control (Play, Pause, Next, …) is Premium-only; Spotify
	// reports this as HTTP 403.
	ErrPremiumRequired = errors.New("spotify: premium account required")
)

func (c *Client) Devices(ctx context.Context, userID string) ([]Device, error) {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	devices, err := sc.PlayerDevices(ctx)
	if err != nil {
		return nil, wrapError("devices", err)
	}
	result := make([]Device, len(devices))
	for i, d := range devices {
		result[i] = deviceFrom(d)
	}
	return result, nil
}

// CurrentPlayback reports what the user is currently playing, or nil when no
// device is active. Use it to answer "what's playing right now?".
func (c *Client) CurrentPlayback(ctx context.Context, userID string) (*Playback, error) {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Market relinks the returned track to one playable in the user's region.
	state, err := sc.PlayerState(ctx, spotify.Market(spotify.MarketFromToken))
	if err != nil {
		return nil, wrapError("current playback", err)
	}
	// Spotify answers with 204 No Content when nothing is active, which zmb3
	// surfaces as a zero-valued PlayerState; the empty device ID is the tell.
	if state.Device.ID == "" {
		return nil, nil
	}
	pb := &Playback{
		Device:      deviceFrom(state.Device),
		IsPlaying:   state.Playing,
		ProgressMs:  int(state.Progress),
		ContextURI:  string(state.PlaybackContext.URI),
		ContextType: state.PlaybackContext.Type,
	}
	if state.Item != nil {
		track := trackFrom(*state.Item)
		pb.Track = &track
	}
	return pb, nil
}

// Play starts or resumes playback on the user's device. uri selects what to
// play and is routed by its Spotify URI type: a track URI plays that single
// track, while an album, playlist, or artist URI plays that context. An empty
// uri resumes whatever is already loaded. An empty deviceID targets the user's
// currently active device.
func (c *Client) Play(ctx context.Context, userID, deviceID, uri string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	opts := &spotify.PlayOptions{}
	if deviceID != "" {
		id := spotify.ID(deviceID)
		opts.DeviceID = &id
	}
	if uri != "" {
		if isContextURI(uri) {
			ctxURI := spotify.URI(uri)
			opts.PlaybackContext = &ctxURI
		} else {
			opts.URIs = []spotify.URI{spotify.URI(uri)}
		}
	}
	if err := sc.PlayOpt(ctx, opts); err != nil {
		return wrapError("play", err)
	}
	return nil
}

func (c *Client) Pause(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Pause(ctx); err != nil {
		return wrapError("pause", err)
	}
	return nil
}

func (c *Client) Resume(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Play(ctx); err != nil {
		return wrapError("resume", err)
	}
	return nil
}

// Next skips to the next track in the user's queue.
func (c *Client) Next(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Next(ctx); err != nil {
		return wrapError("next", err)
	}
	return nil
}

// Previous skips to the previous track in the user's queue.
func (c *Client) Previous(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Previous(ctx); err != nil {
		return wrapError("previous", err)
	}
	return nil
}

// Seek moves the currently playing track to positionMs milliseconds from its
// start. A position past the track's end advances to the next track.
func (c *Client) Seek(ctx context.Context, userID string, positionMs int) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Seek(ctx, positionMs); err != nil {
		return wrapError("seek", err)
	}
	return nil
}

func (c *Client) SetVolume(ctx context.Context, userID string, percent int) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Volume(ctx, percent); err != nil {
		return wrapError("set volume", err)
	}
	return nil
}

func deviceFrom(d spotify.PlayerDevice) Device {
	return Device{
		ID:       d.ID.String(),
		Name:     d.Name,
		Type:     d.Type,
		IsActive: d.Active,
		Volume:   int(d.Volume),
	}
}

// isContextURI reports whether uri names a Spotify context that playback can be
// pointed at as a whole — an album, artist, or playlist — as opposed to a
// single track. URIs have the form "spotify:<type>:<id>".
func isContextURI(uri string) bool {
	parts := strings.SplitN(uri, ":", 3)
	if len(parts) < 2 {
		return false
	}
	switch parts[1] {
	case "album", "artist", "playlist":
		return true
	default:
		return false
	}
}
