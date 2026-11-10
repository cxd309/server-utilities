package pathutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithin(t *testing.T) {
	tests := []struct {
		parent, path string
		want         bool
	}{
		{"/mnt/prudence/music", "/mnt/prudence/music", true},
		{"/mnt/prudence/music", "/mnt/prudence/music/mirror", true},
		{"/mnt/prudence/music", "/mnt/prudence/music-mirror", false}, // sibling, not child
		{"/mnt/prudence/music", "/mnt/prudence", false},
		{"/mnt/prudence/music", "/mnt/prudence/music/..odd name", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, Within(tt.parent, tt.path), "within(%q, %q)", tt.parent, tt.path)
	}
}
