package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cxd309/server-utilities/internal/musicbrainz"
)

func rels(ids ...string) []musicbrainz.Release {
	out := make([]musicbrainz.Release, len(ids))
	for i, id := range ids {
		out[i] = musicbrainz.Release{ID: id, Title: "25"}
	}
	return out
}

func TestAutoPick(t *testing.T) {
	i, ok := autoPick(rels("a", "b", "c"), map[string]bool{"b": true})
	assert.True(t, ok)
	assert.Equal(t, 1, i, "the only edition containing the disc")

	i, ok = autoPick(rels("a"), nil)
	assert.True(t, ok)
	assert.Equal(t, 0, i, "the only edition")

	_, ok = autoPick(rels("a", "b"), map[string]bool{"a": true, "b": true})
	assert.False(t, ok, "several editions contain the disc")

	_, ok = autoPick(rels("a", "b"), nil)
	assert.False(t, ok, "no disc information")
}

func TestPromptPick(t *testing.T) {
	var out bytes.Buffer
	i, err := promptPick(strings.NewReader("x\n9\n2\n"), &out, 3)
	require.NoError(t, err)
	assert.Equal(t, 1, i)
	assert.Equal(t, 3, strings.Count(out.String(), "Pick 1-3"), "re-asks after bad input")

	i, err = promptPick(strings.NewReader("0\n"), &out, 3)
	require.NoError(t, err)
	assert.Equal(t, -1, i, "0 means none")

	_, err = promptPick(strings.NewReader(""), &out, 3)
	assert.Error(t, err, "EOF without a choice")
}

func TestPrintReleases(t *testing.T) {
	var out bytes.Buffer
	r := rels("a", "b")
	r[1].Country = "GB"
	printReleases(&out, r, map[string]bool{"b": true})

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 3)
	assert.Contains(t, lines[0], "TITLE")
	assert.Contains(t, lines[2], "GB")
	assert.Contains(t, lines[2], "✓ this disc")
	assert.NotContains(t, lines[1], "✓")
}
