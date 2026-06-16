package spotify

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func TestMissingScopes(t *testing.T) {
	tests := []struct {
		name     string
		required []string
		granted  []string
		want     []string
	}{
		{
			name:     "all granted",
			required: RequiredScopes,
			granted:  RequiredScopes,
			want:     nil,
		},
		{
			name:     "none granted",
			required: []string{"a", "b"},
			granted:  nil,
			want:     []string{"a", "b"},
		},
		{
			name:     "partial grant returns only the gap",
			required: []string{"a", "b", "c"},
			granted:  []string{"b", "x"},
			want:     []string{"a", "c"},
		},
		{
			name:     "extra granted scopes are ignored",
			required: []string{"a"},
			granted:  []string{"a", "b", "c"},
			want:     nil,
		},
		{
			name:     "order of granted does not matter",
			required: []string{"a", "b"},
			granted:  []string{"b", "a"},
			want:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := missingScopes(tt.required, tt.granted)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("missingScopes(%v, %v) = %v, want %v", tt.required, tt.granted, got, tt.want)
			}
		})
	}
}

func TestGrantedScopes(t *testing.T) {
	tests := []struct {
		name  string
		extra map[string]interface{}
		want  []string
	}{
		{
			name:  "space separated scope string",
			extra: map[string]interface{}{"scope": "user-read-playback-state user-modify-playback-state"},
			want:  []string{"user-read-playback-state", "user-modify-playback-state"},
		},
		{
			name:  "single scope",
			extra: map[string]interface{}{"scope": "playlist-read-private"},
			want:  []string{"playlist-read-private"},
		},
		{
			name:  "empty scope string",
			extra: map[string]interface{}{"scope": ""},
			want:  []string{},
		},
		{
			name:  "extra whitespace is collapsed",
			extra: map[string]interface{}{"scope": "  a   b  "},
			want:  []string{"a", "b"},
		},
		{
			name:  "no scope field",
			extra: map[string]interface{}{},
			want:  []string{},
		},
		{
			name:  "scope field of unexpected type",
			extra: map[string]interface{}{"scope": 42},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := (&oauth2.Token{}).WithExtra(tt.extra)
			got := grantedScopes(token)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grantedScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScopeErrorWrapsSentinel(t *testing.T) {
	err := error(&ScopeError{
		Granted: []string{"a"},
		Missing: []string{"b", "c"},
	})

	if !errors.Is(err, ErrMissingScopes) {
		t.Errorf("errors.Is(err, ErrMissingScopes) = false, want true")
	}

	var se *ScopeError
	if !errors.As(err, &se) {
		t.Fatalf("errors.As(err, *ScopeError) = false, want true")
	}
	if !reflect.DeepEqual(se.Missing, []string{"b", "c"}) {
		t.Errorf("se.Missing = %v, want [b c]", se.Missing)
	}
}

func TestSentinelFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "404 no active device",
			err:  spotify.Error{Status: 404, Message: "Player command failed: No active device found"},
			want: ErrNoActiveDevice,
		},
		{
			name: "404 other reason is not mapped",
			err:  spotify.Error{Status: 404, Message: "Not found"},
			want: nil,
		},
		{
			name: "403 premium required",
			err:  spotify.Error{Status: 403, Message: "Player command failed: Premium required"},
			want: ErrPremiumRequired,
		},
		{
			name: "403 other reason is not mapped",
			err:  spotify.Error{Status: 403, Message: "Forbidden"},
			want: nil,
		},
		{
			name: "429 rate limited",
			err:  spotify.Error{Status: 429, Message: "API rate limit exceeded"},
			want: ErrRateLimited,
		},
		{
			name: "unrecognized status",
			err:  spotify.Error{Status: 500, Message: "Internal server error"},
			want: nil,
		},
		{
			name: "non-API error",
			err:  errors.New("boom"),
			want: nil,
		},
		{
			name: "wrapped API error is still matched",
			err:  fmt.Errorf("play: %w", spotify.Error{Status: 429, Message: "slow down"}),
			want: ErrRateLimited,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sentinelFor(tt.err); got != tt.want {
				t.Errorf("sentinelFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	t.Run("nil passes through", func(t *testing.T) {
		if got := wrapError("play", nil); got != nil {
			t.Errorf("wrapError(nil) = %v, want nil", got)
		}
	})

	t.Run("joins sentinel and keeps original in the chain", func(t *testing.T) {
		api := spotify.Error{Status: 404, Message: "Player command failed: No active device found"}
		err := wrapError("play", api)

		if !errors.Is(err, ErrNoActiveDevice) {
			t.Errorf("errors.Is(err, ErrNoActiveDevice) = false, want true")
		}
		var apiErr spotify.Error
		if !errors.As(err, &apiErr) {
			t.Errorf("errors.As(err, *spotify.Error) = false, want true")
		}
	})

	t.Run("plain error is annotated without a sentinel", func(t *testing.T) {
		base := errors.New("boom")
		err := wrapError("seek", base)

		if !errors.Is(err, base) {
			t.Errorf("errors.Is(err, base) = false, want true")
		}
		for _, sentinel := range []error{ErrNoActiveDevice, ErrPremiumRequired, ErrRateLimited} {
			if errors.Is(err, sentinel) {
				t.Errorf("errors.Is(err, %v) = true, want false", sentinel)
			}
		}
	})
}
