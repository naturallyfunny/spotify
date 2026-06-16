package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"spotify-api/db"
	"spotify-api/spotify"
	"sync"

	"go.naturallyfunny.dev/api"
	"go.naturallyfunny.dev/api/identity"
)

// PlayRequest defines the expected JSON body for POST /play.
type PlayRequest struct {
	DeviceID string `json:"device_id"`
	PlayMu   string `json:"play_mu"`  // Track URI to play (e.g., "spotify:track:xxx")
	Command  string `json:"command"`  // "pause", "resume"
	Volume   *int   `json:"volume"`   // Volume percent (0-100), pointer to distinguish 0 from absent
}

// PlayResponse defines the JSON response for POST /play.
type PlayResponse struct {
	Success bool     `json:"success"`
	Output  []string `json:"output"`
	Error   string   `json:"error,omitempty"`
}

// PlayHandler handles POST /play.
func PlayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, api.NewError(api.InvalidArgument, "method not allowed, use POST"))
		return
	}

	var req PlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, api.NewError(api.InvalidArgument, "invalid JSON body: "+err.Error()))
		return
	}
	defer r.Body.Close()

	userID, err := identity.GetUserIDFromContext(r.Context())
	if err != nil || userID == "" {
		api.WriteError(w, api.NewError(api.Unauthenticated, "x-user-id is required or invalid context"))
		return
	}

	refreshToken, err := db.GetRefreshToken(userID)
	if err != nil {
		api.WriteError(w, api.NewError(api.Unauthenticated, "failed to get refresh token: "+err.Error()))
		return
	}

	// Exchange refresh token for access token
	accessToken, err := spotify.ExchangeRefreshToken(refreshToken)
	if err != nil {
		api.WriteError(w, api.NewError(api.Unauthenticated, "token exchange failed: "+err.Error()))
		return
	}

	// If device_id is empty, fetch the first available device
	deviceID := req.DeviceID
	if deviceID == "" {
		devices, err := spotify.GetDevices(accessToken)
		if err != nil {
			api.WriteError(w, api.NewError(api.Internal, "failed to get devices: "+err.Error()))
			return
		}
		if len(devices) == 0 {
			api.WriteError(w, api.NewError(api.NotFound, "no active Spotify devices found"))
			return
		}
		deviceID = devices[0].ID
	}

	// Dispatch concurrent tasks for play_mu, command, and volume
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		output  []string
		errList []string
	)

	appendOutput := func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		output = append(output, msg)
	}

	appendError := func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		errList = append(errList, msg)
	}

	// Task 1: play_mu — play a specific track
	if req.PlayMu != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			uris := []string{req.PlayMu}
			if err := spotify.PlayTrack(accessToken, deviceID, uris); err != nil {
				appendError(fmt.Sprintf("play_mu failed: %v", err))
			} else {
				appendOutput(fmt.Sprintf("Now playing: %s", req.PlayMu))
			}
		}()
	}

	// Task 2: command — pause or resume
	if req.Command != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			switch req.Command {
			case "pause":
				if err := spotify.Pause(accessToken, deviceID); err != nil {
					appendError(fmt.Sprintf("pause failed: %v", err))
				} else {
					appendOutput("Playback paused")
				}
			case "resume":
				if err := spotify.Resume(accessToken, deviceID); err != nil {
					appendError(fmt.Sprintf("resume failed: %v", err))
				} else {
					appendOutput("Playback resumed")
				}
			default:
				appendError(fmt.Sprintf("unknown command: %s (supported: pause, resume)", req.Command))
			}
		}()
	}

	// Task 3: volume
	if req.Volume != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vol := *req.Volume
			if vol < 0 || vol > 100 {
				appendError(fmt.Sprintf("volume must be between 0 and 100, got %d", vol))
				return
			}
			if err := spotify.SetVolume(accessToken, deviceID, vol); err != nil {
				appendError(fmt.Sprintf("set volume failed: %v", err))
			} else {
				appendOutput(fmt.Sprintf("Volume set to %d%%", vol))
			}
		}()
	}

	wg.Wait()

	// Combine output
	allOutput := append(output, errList...)

	resp := PlayResponse{
		Success: true,
		Output:  allOutput,
	}

	if len(errList) > 0 && len(output) == 0 {
		resp.Success = false
		resp.Error = "all operations failed"
		api.WriteSuccess(w, api.Internal, "All operations failed", resp, nil)
		return
	}

	if len(errList) > 0 {
		resp.Error = "some operations failed"
		api.WriteSuccess(w, api.OK, "Play command processed with some errors", resp, nil)
		return
	}

	api.WriteSuccess(w, api.OK, "Play command processed successfully", resp, nil)
}
