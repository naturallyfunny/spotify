package spotify

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// --- Playlist Types ---

// SimplifiedPlaylist represents a user's playlist (simplified view).
type SimplifiedPlaylist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TracksTotal struct {
		Total int `json:"total"`
	} `json:"tracks"`
	ExternalURLs map[string]string `json:"external_urls"`
}

// PlaylistsResponse wraps the paginated playlists response.
type PlaylistsResponse struct {
	Items []SimplifiedPlaylist `json:"items"`
	Total int                  `json:"total"`
}

// --- Genre Seeds ---

// GenreSeedsResponse wraps the available genre seeds response.
type GenreSeedsResponse struct {
	Genres []string `json:"genres"`
}

// --- Search Types ---

// TrackItem represents a track from search results.
type TrackItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URI         string `json:"uri"`
	DurationMs  int    `json:"duration_ms"`
	Artists     []struct {
		Name string `json:"name"`
	} `json:"artists"`
	Album struct {
		Name string `json:"name"`
	} `json:"album"`
}

// SearchTracksResponse wraps the search results for tracks.
type SearchTracksResponse struct {
	Tracks struct {
		Items []TrackItem `json:"items"`
		Total int         `json:"total"`
	} `json:"tracks"`
}

// SearchPlaylistsResponse wraps the search results for playlists.
type SearchPlaylistsResponse struct {
	Playlists struct {
		Items []SimplifiedPlaylist `json:"items"`
		Total int                  `json:"total"`
	} `json:"playlists"`
}

// --- Playlist Tracks ---

// PlaylistTrackItem wraps a track inside a playlist.
type PlaylistTrackItem struct {
	Track TrackItem `json:"track"`
}

// PlaylistTracksResponse wraps the paginated playlist tracks response.
type PlaylistTracksResponse struct {
	Items []PlaylistTrackItem `json:"items"`
	Total int                 `json:"total"`
}

// --- Recommendations ---

// RecommendationsResponse wraps the recommendations response.
type RecommendationsResponse struct {
	Tracks []TrackItem `json:"tracks"`
}

// =============================================================================
// API Functions
// =============================================================================

// GetPlaylists retrieves the current user's playlists.
func GetPlaylists(accessToken string) (*PlaylistsResponse, error) {
	body, err := SpotifyGet(accessToken, "/me/playlists?limit=20")
	if err != nil {
		return nil, fmt.Errorf("GetPlaylists: %w", err)
	}

	var resp PlaylistsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("GetPlaylists: failed to decode: %w", err)
	}

	return &resp, nil
}

// GetGenreSeeds retrieves the available genre seeds for recommendations.
func GetGenreSeeds(accessToken string) ([]string, error) {
	body, err := SpotifyGet(accessToken, "/recommendations/available-genre-seeds")
	if err != nil {
		return nil, fmt.Errorf("GetGenreSeeds: %w", err)
	}

	var resp GenreSeedsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("GetGenreSeeds: failed to decode: %w", err)
	}

	return resp.Genres, nil
}

// SearchTrack searches for tracks matching the query.
func SearchTrack(accessToken, query string) (*SearchTracksResponse, error) {
	endpoint := fmt.Sprintf("/search?q=%s&type=track&limit=10", url.QueryEscape(query))
	body, err := SpotifyGet(accessToken, endpoint)
	if err != nil {
		return nil, fmt.Errorf("SearchTrack: %w", err)
	}

	var resp SearchTracksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("SearchTrack: failed to decode: %w", err)
	}

	return &resp, nil
}

// SearchPlaylist searches for playlists matching the query.
func SearchPlaylist(accessToken, query string) (*SearchPlaylistsResponse, error) {
	endpoint := fmt.Sprintf("/search?q=%s&type=playlist&limit=10", url.QueryEscape(query))
	body, err := SpotifyGet(accessToken, endpoint)
	if err != nil {
		return nil, fmt.Errorf("SearchPlaylist: %w", err)
	}

	var resp SearchPlaylistsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("SearchPlaylist: failed to decode: %w", err)
	}

	return &resp, nil
}

// GetPlaylistTracks retrieves the tracks from a specific playlist.
func GetPlaylistTracks(accessToken, playlistID string) (*PlaylistTracksResponse, error) {
	endpoint := fmt.Sprintf("/playlists/%s/tracks?limit=50", url.QueryEscape(playlistID))
	body, err := SpotifyGet(accessToken, endpoint)
	if err != nil {
		return nil, fmt.Errorf("GetPlaylistTracks: %w", err)
	}

	var resp PlaylistTracksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("GetPlaylistTracks: failed to decode: %w", err)
	}

	return &resp, nil
}

// GetRecommendations retrieves track recommendations based on genre seeds.
func GetRecommendations(accessToken string, genres []string) (*RecommendationsResponse, error) {
	seedGenres := url.QueryEscape(joinStrings(genres, ","))
	endpoint := fmt.Sprintf("/recommendations?seed_genres=%s&limit=20", seedGenres)
	body, err := SpotifyGet(accessToken, endpoint)
	if err != nil {
		return nil, fmt.Errorf("GetRecommendations: %w", err)
	}

	var resp RecommendationsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("GetRecommendations: failed to decode: %w", err)
	}

	return &resp, nil
}

// joinStrings joins a slice of strings with a separator (avoids importing strings package just for this).
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
