package spotify

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Device represents a Spotify playback device.
type Device struct {
	ID            string `json:"id"`
	IsActive      bool   `json:"is_active"`
	IsRestricted  bool   `json:"is_restricted"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	VolumePercent int    `json:"volume_percent"`
}

// DevicesResponse is the wrapper for the /me/player/devices response.
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// GetDevices retrieves the list of available playback devices.
func GetDevices(accessToken string) ([]Device, error) {
	body, err := SpotifyGet(accessToken, "/me/player/devices")
	if err != nil {
		return nil, fmt.Errorf("GetDevices: %w", err)
	}

	var resp DevicesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("GetDevices: failed to decode: %w", err)
	}

	return resp.Devices, nil
}

// PlayTrack starts playback of a specific track (or list of URIs) on a device.
// If uris is nil, it simply resumes playback.
func PlayTrack(accessToken, deviceID string, uris []string) error {
	endpoint := "/me/player/play"
	if deviceID != "" {
		endpoint += "?device_id=" + url.QueryEscape(deviceID)
	}

	var payload []byte
	if len(uris) > 0 {
		data := map[string]interface{}{
			"uris": uris,
		}
		var err error
		payload, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("PlayTrack: failed to marshal: %w", err)
		}
	}

	_, err := SpotifyPut(accessToken, endpoint, payload)
	if err != nil {
		return fmt.Errorf("PlayTrack: %w", err)
	}

	return nil
}

// Resume resumes playback on a device (play without URIs).
func Resume(accessToken, deviceID string) error {
	return PlayTrack(accessToken, deviceID, nil)
}

// Pause pauses playback on a device.
func Pause(accessToken, deviceID string) error {
	endpoint := "/me/player/pause"
	if deviceID != "" {
		endpoint += "?device_id=" + url.QueryEscape(deviceID)
	}

	_, err := SpotifyPut(accessToken, endpoint, nil)
	if err != nil {
		return fmt.Errorf("Pause: %w", err)
	}

	return nil
}

// SetVolume sets the playback volume on a device.
func SetVolume(accessToken, deviceID string, volumePercent int) error {
	endpoint := fmt.Sprintf("/me/player/volume?volume_percent=%d", volumePercent)
	if deviceID != "" {
		endpoint += "&device_id=" + url.QueryEscape(deviceID)
	}

	_, err := SpotifyPut(accessToken, endpoint, nil)
	if err != nil {
		return fmt.Errorf("SetVolume: %w", err)
	}

	return nil
}
