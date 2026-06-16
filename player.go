package spotify

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func (c *Client) Devices(ctx context.Context, userID string) ([]Device, error) {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	devices, err := sc.PlayerDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("devices: %w", err)
	}
	result := make([]Device, len(devices))
	for i, d := range devices {
		result[i] = Device{
			ID:       d.ID.String(),
			Name:     d.Name,
			Type:     d.Type,
			IsActive: d.Active,
			Volume:   int(d.Volume),
		}
	}
	return result, nil
}

func (c *Client) Play(ctx context.Context, userID, deviceID, trackURI string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	opts := &spotify.PlayOptions{}
	if deviceID != "" {
		id := spotify.ID(deviceID)
		opts.DeviceID = &id
	}
	if trackURI != "" {
		opts.URIs = []spotify.URI{spotify.URI(trackURI)}
	}
	if err := sc.PlayOpt(ctx, opts); err != nil {
		return fmt.Errorf("play: %w", err)
	}
	return nil
}

func (c *Client) Pause(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Pause(ctx); err != nil {
		return fmt.Errorf("pause: %w", err)
	}
	return nil
}

func (c *Client) Resume(ctx context.Context, userID string) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Play(ctx); err != nil {
		return fmt.Errorf("resume: %w", err)
	}
	return nil
}

func (c *Client) SetVolume(ctx context.Context, userID string, percent int) error {
	sc, err := c.clientFor(ctx, userID)
	if err != nil {
		return err
	}
	if err := sc.Volume(ctx, percent); err != nil {
		return fmt.Errorf("set volume: %w", err)
	}
	return nil
}
