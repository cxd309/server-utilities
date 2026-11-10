package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// prune deletes files in dest that no longer match a job,
// plus leftovers from interrupted runs, then removes any folders left empty.
// It only touches files music-mirror could have written itself.
// With dryRun it reports what it would remove and changes nothing.
func prune(dest string, jobs []job, dryRun bool) (removed []string, err error) {
	if _, err = os.Stat(dest); errors.Is(err, fs.ErrNotExist) {
		return nil, nil // nothing mirrored yet
	}

	keep := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		keep[j.dst] = true
	}

	var dirs []string
	err = filepath.WalkDir(dest, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path != dest {
				dirs = append(dirs, path)
			}
			return nil
		}
		if keep[path] || !isOurs(path) {
			return nil
		}
		removed = append(removed, path)
		if dryRun {
			return nil
		}
		return os.Remove(path)
	})
	if err != nil || dryRun {
		return removed, err
	}

	// WalkDir lists parents before children,
	// so going backwards removes the deepest folders first and their parents can then empty out too.
	for i := len(dirs) - 1; i >= 0; i-- {
		_ = os.Remove(dirs[i]) // only succeeds on an empty folder
	}
	return removed, nil
}

// isOurs reports whether path looks like something music-mirror writes:
// a finished mp3, or a temporary file from an interrupted run.
func isOurs(path string) (ours bool) {
	name := strings.ToLower(filepath.Base(path))
	for _, suffix := range []string{".mp3", ".mp3.part", ".mp3.part-id3v2"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
