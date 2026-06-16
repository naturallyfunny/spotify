package spotify

import (
	"reflect"
	"testing"

	"github.com/zmb3/spotify/v2"
)

func TestTrackFrom(t *testing.T) {
	full := spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			ID:   "6rqhFgbbKwnb9MLmUQDhG6",
			Name: "Bohemian Rhapsody",
			Artists: []spotify.SimpleArtist{
				{Name: "Queen"},
				{Name: "Someone Else"},
			},
			URI:          "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
			ExternalURLs: map[string]string{"spotify": "https://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6"},
		},
	}

	want := Track{
		ID:      "6rqhFgbbKwnb9MLmUQDhG6",
		Name:    "Bohemian Rhapsody",
		Artists: []string{"Queen", "Someone Else"},
		URI:     "spotify:track:6rqhFgbbKwnb9MLmUQDhG6",
		URL:     "https://open.spotify.com/track/6rqhFgbbKwnb9MLmUQDhG6",
	}

	if got := trackFrom(full); !reflect.DeepEqual(got, want) {
		t.Errorf("trackFrom() = %+v, want %+v", got, want)
	}
}

func TestTrackFromNoArtists(t *testing.T) {
	got := trackFrom(spotify.FullTrack{SimpleTrack: spotify.SimpleTrack{ID: "x", Name: "y"}})
	if len(got.Artists) != 0 {
		t.Errorf("Artists = %v, want empty", got.Artists)
	}
}

func TestPlaylistFrom(t *testing.T) {
	simple := spotify.SimplePlaylist{
		ID:           "37i9dQZF1DXcBWIGoYBM5M",
		Name:         "Today's Top Hits",
		Description:  "The hottest tracks.",
		Tracks:       spotify.PlaylistTracks{Total: 50},
		ExternalURLs: map[string]string{"spotify": "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M"},
	}

	want := Playlist{
		ID:          "37i9dQZF1DXcBWIGoYBM5M",
		Name:        "Today's Top Hits",
		Description: "The hottest tracks.",
		Total:       50,
		URL:         "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
	}

	if got := playlistFrom(simple); !reflect.DeepEqual(got, want) {
		t.Errorf("playlistFrom() = %+v, want %+v", got, want)
	}
}
