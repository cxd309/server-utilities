package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// stateFile sits in the mirror root and records when music-sync last finished.
// It's hidden, so music-mirror's scan and prune ignore it.
const stateFile = ".music-sync"

// trackExts are the files music-mirror produces: ALAC from FLAC, MP3 copied as-is.
var trackExts = map[string]bool{".m4a": true, ".mp3": true}

// listMirror returns every track (.m4a or .mp3) under dir, skipping hidden
// files and folders. music-mirror's .part files don't match trackExts either.
func listMirror(dir string) (files []mirrorFile, err error) {
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != dir && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() || !trackExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		info, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}
		files = append(files, mirrorFile{path: path, modTime: info.ModTime()})
		return nil
	})
	return files, err
}

// readLastSync returns when music-sync last finished against mirror,
// or the zero time if it never has.
func readLastSync(mirror string) (last time.Time, err error) {
	var data []byte
	if data, err = os.ReadFile(filepath.Join(mirror, stateFile)); errors.Is(err, fs.ErrNotExist) {
		return time.Time{}, nil
	} else if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data)))
}

// writeLastSync records t as the time of the last successful sync.
func writeLastSync(mirror string, t time.Time) (err error) {
	return os.WriteFile(filepath.Join(mirror, stateFile), []byte(t.Format(time.RFC3339Nano)+"\n"), 0o644)
}
