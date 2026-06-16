package spotify

import "testing"

func TestIsContextURI(t *testing.T) {
	tests := []struct {
		uri  string
		want bool
	}{
		{"spotify:track:6rqhFgbbKwnb9MLmUQDhG6", false},
		{"spotify:album:1DFixLWuPkv3KT3TnV35m3", true},
		{"spotify:playlist:37i9dQZF1DXcBWIGoYBM5M", true},
		{"spotify:artist:0OdUWJ0sBjDrqHygGUXeCF", true},
		{"spotify:episode:512ojhOuo1ktJprKbVcKyQ", false},
		{"", false},
		{"not-a-uri", false},
		{"spotify:track", false},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got := isContextURI(tt.uri); got != tt.want {
				t.Errorf("isContextURI(%q) = %v, want %v", tt.uri, got, tt.want)
			}
		})
	}
}
