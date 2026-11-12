package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLowerKeys(t *testing.T) {
	got := lowerKeys(map[string]string{
		"title":    "A Wizard of Earthsea",
		"SUBTITLE": "The First Book of Earthsea",
		"Album":    "Mixed Case",
	})
	assert.Equal(t, map[string]string{
		"title":    "A Wizard of Earthsea",
		"subtitle": "The First Book of Earthsea",
		"album":    "Mixed Case",
	}, got)
}

func TestLowerKeysNil(t *testing.T) {
	got := lowerKeys(nil)
	assert.NotNil(t, got, "safe to look keys up in")
	assert.Empty(t, got)
}

func TestTidy(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Chapter  2. The Shadow", "Chapter 2. The Shadow"},
		{"  padded  ", "padded"},
		{"tab\tand\nnewline", "tab and newline"},
		{"   ", ""},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tidy(tt.in), "tidy(%q)", tt.in)
	}
}
