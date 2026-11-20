package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// prune deletes book folders in dest that no longer match a book in the library,
// plus build folders left by interrupted runs, then removes author folders left empty.
// It only deletes a book folder holding nothing but .m4b parts (and hidden files like .DS_Store),
// so anything else put in the mirror by hand is left alone.
// With dryRun it reports what it would remove and changes nothing
func prune(dest string, books []book, dryRun bool) (removed []string, err error) {
	var authors []os.DirEntry
	if authors, err = os.ReadDir(dest); errors.Is(err, fs.ErrNotExist) {
		return nil, nil // nothing mirrored yet
	} else if err != nil {
		return nil, err
	}

	keep := make(map[string]bool, len(books))
	for _, b := range books {
		keep[b.dst] = true
	}

	for _, a := range authors {
		if !a.IsDir() || hidden(a.Name()) {
			continue
		}
		authorPath := filepath.Join(dest, a.Name())

		var titles []os.DirEntry
		if titles, err = os.ReadDir(authorPath); err != nil {
			return removed, err
		}
		for _, t := range titles {
			titlePath := filepath.Join(authorPath, t.Name())
			if !t.IsDir() || keep[titlePath] {
				continue
			}
			if hidden(t.Name()) && !isLeftover(t.Name()) {
				continue
			}
			if !hidden(t.Name()) {
				var ours bool
				if ours, err = onlyParts(titlePath); err != nil {
					return removed, err
				}
				if !ours {
					continue
				}
			}

			removed = append(removed, titlePath)
			if dryRun {
				continue
			}
			if err = os.RemoveAll(titlePath); err != nil {
				return removed, err
			}
		}

		if !dryRun {
			_ = os.Remove(authorPath) // only succeeds on an empty folder
		}
	}
	return removed, nil
}

// isLeftover reports whether name is a hidden folder from buildBook or swapDir
// that an interrupted run never cleaned up
func isLeftover(name string) (yes bool) {
	return hidden(name) && (strings.HasSuffix(name, ".building") || strings.HasSuffix(name, ".old"))
}

// onlyParts reports whether dir holds nothing but .m4b files and hidden files,
// meaning everything in it was written by audiobook-mirror (or the Mac)
func onlyParts(dir string) (ours bool, err error) {
	var entries []os.DirEntry
	if entries, err = os.ReadDir(dir); err != nil {
		return false, err
	}
	for _, e := range entries {
		if hidden(e.Name()) {
			continue
		}
		if e.IsDir() || strings.ToLower(filepath.Ext(e.Name())) != ".m4b" {
			return false, nil
		}
	}
	return true, nil
}
