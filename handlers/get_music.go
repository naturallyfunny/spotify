package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"go.avagenc.com/spotify/db"
	"go.avagenc.com/spotify/spotify"
	"sync"

	"go.naturallyfunny.dev/api"
	"go.naturallyfunny.dev/api/identity"
)

// GetMusicRequest defines the expected JSON body for GET /get-music.
type GetMusicRequest struct {
	Query             string   `json:"query"`              // Search query for tracks
	PlaylistQuery     string   `json:"playlist_query"`     // Search query for playlists
	PlaylistTrack     string   `json:"playlist_track"`     // Playlist ID to fetch tracks from
	GenreRecomendation []string `json:"genre_recomendation"` // Genre seeds for recommendations
}

// GetMusicResponse defines the JSON response for GET /get-music.
type GetMusicResponse struct {
	Success      bool                        `json:"success"`
	Devices      []spotify.Device            `json:"devices,omitempty"`
	UserPlaylist *spotify.PlaylistsResponse  `json:"user_playlist,omitempty"`
	GenreSeed    []string                    `json:"genre_seed,omitempty"`
	Output       []string                    `json:"output,omitempty"`
	Error        string                      `json:"error,omitempty"`

	// Optional fields — populated if corresponding request params are provided
	SearchResults   *spotify.SearchTracksResponse    `json:"search_results,omitempty"`
	PlaylistSearch  *spotify.SearchPlaylistsResponse `json:"playlist_search,omitempty"`
	PlaylistTracks  *spotify.PlaylistTracksResponse  `json:"playlist_tracks,omitempty"`
	Recommendations *spotify.RecommendationsResponse `json:"recommendations,omitempty"`
}

// GetMusicHandler handles GET /get-music.
func GetMusicHandler(w http.ResponseWriter, r *http.Request) {
	// We accept both GET and POST for flexibility
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		api.WriteError(w, api.NewError(api.InvalidArgument, "method not allowed, use GET or POST"))
		return
	}

	var req GetMusicRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			api.WriteError(w, api.NewError(api.InvalidArgument, "invalid JSON body: "+err.Error()))
			return
		}
		defer r.Body.Close()
	}

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

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		output  []string
		errList []string
		resp    GetMusicResponse
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

	// =========================================================================
	// Mandatory Tasks (always run concurrently)
	// =========================================================================

	// Task 1: Get Devices
	wg.Add(1)
	go func() {
		defer wg.Done()
		devices, err := spotify.GetDevices(accessToken)
		if err != nil {
			appendError(fmt.Sprintf("get devices failed: %v", err))
			return
		}
		mu.Lock()
		resp.Devices = devices
		mu.Unlock()
		appendOutput(fmt.Sprintf("Found %d device(s)", len(devices)))
	}()

	// Task 2: Get User Playlists
	wg.Add(1)
	go func() {
		defer wg.Done()
		playlists, err := spotify.GetPlaylists(accessToken)
		if err != nil {
			appendError(fmt.Sprintf("get playlists failed: %v", err))
			return
		}
		mu.Lock()
		resp.UserPlaylist = playlists
		mu.Unlock()
		appendOutput(fmt.Sprintf("Found %d playlist(s)", playlists.Total))
	}()

	// Task 3: Get Genre Seeds
	wg.Add(1)
	go func() {
		defer wg.Done()
		genres, err := spotify.GetGenreSeeds(accessToken)
		if err != nil {
			appendError(fmt.Sprintf("get genre seeds failed: %v", err))
			return
		}
		mu.Lock()
		resp.GenreSeed = genres
		mu.Unlock()
		appendOutput(fmt.Sprintf("Found %d genre seed(s)", len(genres)))
	}()

	// =========================================================================
	// Optional Tasks (run only if corresponding parameters are provided)
	// =========================================================================

	// Task 4: Search Track (if query is provided)
	if req.Query != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := spotify.SearchTrack(accessToken, req.Query)
			if err != nil {
				appendError(fmt.Sprintf("search track failed: %v", err))
				return
			}
			mu.Lock()
			resp.SearchResults = results
			mu.Unlock()
			appendOutput(fmt.Sprintf("Track search for '%s': %d result(s)", req.Query, results.Tracks.Total))
		}()
	}

	// Task 5: Search Playlist (if playlist_query is provided)
	if req.PlaylistQuery != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := spotify.SearchPlaylist(accessToken, req.PlaylistQuery)
			if err != nil {
				appendError(fmt.Sprintf("search playlist failed: %v", err))
				return
			}
			mu.Lock()
			resp.PlaylistSearch = results
			mu.Unlock()
			appendOutput(fmt.Sprintf("Playlist search for '%s': %d result(s)", req.PlaylistQuery, results.Playlists.Total))
		}()
	}

	// Task 6: Get Playlist Tracks (if playlist_track is provided)
	if req.PlaylistTrack != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracks, err := spotify.GetPlaylistTracks(accessToken, req.PlaylistTrack)
			if err != nil {
				appendError(fmt.Sprintf("get playlist tracks failed: %v", err))
				return
			}
			mu.Lock()
			resp.PlaylistTracks = tracks
			mu.Unlock()
			appendOutput(fmt.Sprintf("Playlist tracks: %d track(s)", tracks.Total))
		}()
	}

	// Task 7: Get Recommendations (if genre_recomendation is provided)
	if len(req.GenreRecomendation) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recs, err := spotify.GetRecommendations(accessToken, req.GenreRecomendation)
			if err != nil {
				appendError(fmt.Sprintf("get recommendations failed: %v", err))
				return
			}
			mu.Lock()
			resp.Recommendations = recs
			mu.Unlock()
			appendOutput(fmt.Sprintf("Recommendations: %d track(s)", len(recs.Tracks)))
		}()
	}

	wg.Wait()

	// Combine output
	resp.Output = append(output, errList...)
	resp.Success = len(errList) == 0 || len(output) > 0

	if len(errList) > 0 && len(output) == 0 {
		resp.Error = "all operations failed"
		api.WriteSuccess(w, api.Internal, "All operations failed", resp, nil)
		return
	}

	if len(errList) > 0 {
		resp.Error = "some operations failed"
		api.WriteSuccess(w, api.OK, "Get music processed with some errors", resp, nil)
		return
	}

	api.WriteSuccess(w, api.OK, "Get music processed successfully", resp, nil)
}
