package spotify

import (
	"errors"
	"reflect"
	"testing"

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
