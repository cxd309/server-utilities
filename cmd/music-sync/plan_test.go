package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlan(t *testing.T) {
	const mirror = "/Users/chris/Music/music-mirror"
	lastSync := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	before, after := lastSync.Add(-time.Hour), lastSync.Add(time.Hour)

	tracks := []libraryTrack{
		{id: "UNCHANGED", path: mirror + "/A/01.mp3"},
		{id: "RETAGGED", path: mirror + "/A/02.mp3"},
		{id: "PRUNED", path: mirror + "/A/03.mp3"},              // file deleted by music-mirror --prune
		{id: "MISSING", path: ""},                               // Music app shows "!"
		{id: "ELSEWHERE", path: "/Users/chris/Downloads/x.mp3"}, // not ours: leave alone
	}
	files := []mirrorFile{
		{path: mirror + "/A/01.mp3", modTime: before},
		{path: mirror + "/A/02.mp3", modTime: after},
		{path: mirror + "/B/new.mp3", modTime: after},
	}

	assert.Equal(t, syncPlan{
		add:     []string{mirror + "/B/new.mp3"},
		refresh: []string{"RETAGGED"},
		remove:  []string{"MISSING", "PRUNED"},
	}, plan(mirror, tracks, files, lastSync))
}

func TestPlanEmptyLibrary(t *testing.T) {
	const mirror = "/m"
	files := []mirrorFile{{path: "/m/b.mp3"}, {path: "/m/a.mp3"}}

	p := plan(mirror, nil, files, time.Time{})

	assert.Equal(t, []string{"/m/a.mp3", "/m/b.mp3"}, p.add)
	assert.Empty(t, p.refresh, "new tracks are added, not refreshed")
	assert.Empty(t, p.remove)
}

func TestPlanUnicodeNormalisation(t *testing.T) {
	const mirror = "/m"
	decomposed := "/m/Sine\u0301ad.mp3" // e + combining accent, as macOS can report it
	composed := "/m/Sin\u00e9ad.mp3"    // single é, as Go wrote it

	p := plan(mirror, []libraryTrack{{id: "X", path: decomposed}}, []mirrorFile{{path: composed}}, time.Now())

	assert.Equal(t, syncPlan{}, p, "the same file spelled two ways is still one track")
}
