package main

import (
	"slices"
	"time"

	"github.com/cxd309/server-utilities/internal/pathutil"
	"golang.org/x/text/unicode/norm"
)

// libraryTrack is one file track is the Music app's library
type libraryTrack struct {
	id   string // the Music app's persistent ID
	path string // empty when the Music app can't find the file
}

// mirrorFile is one mp3 in the mirror folder
type mirrorFile struct {
	path    string
	modTime time.Time
}

// syncPlan is what needs doing to make the library match the mirror
type syncPlan struct {
	add     []string // mirror paths not yet in the library
	refresh []string // persistent IDs whose file changed since the last sync
	remove  []string // persistent IDs whose file has gone
}

// plan compares the library with the mirror.
// Tracks outside the mirror are left alone unless their file is missing.
func plan(mirror string, tracks []libraryTrack, files []mirrorFile, lastSync time.Time) (p syncPlan) {
	onDisk := make(map[string]bool, len(files))
	for _, f := range files {
		onDisk[norm.NFC.String(f.path)] = true
	}

	inLibrary := make(map[string]string, len(tracks)) // path -> persistent ID
	for _, t := range tracks {
		path := norm.NFC.String(t.path)
		switch {
		case path == "":
			p.remove = append(p.remove, t.id)
		case pathutil.Within(mirror, path) && !onDisk[path]:
			p.remove = append(p.remove, t.id)
		default:
			inLibrary[path] = t.id
		}
	}

	for _, f := range files {
		id, ok := inLibrary[norm.NFC.String(f.path)]
		switch {
		case !ok:
			p.add = append(p.add, f.path)
		case f.modTime.After(lastSync):
			p.refresh = append(p.refresh, id)
		}
	}

	slices.Sort(p.add)
	slices.Sort(p.refresh)
	slices.Sort(p.remove)
	return p
}
