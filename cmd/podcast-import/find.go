package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// findFile returns the path of rel, a slash-separated path under root, matching each name
// in it whatever its Unicode form. An accented letter like the û in "Khazad-dûm" can be stored
// as one character or as a letter plus a combining accent. The two look the same but have
// different bytes, so a name copied from a Mac may not match the name on the server's disk
func findFile(root, rel string) (path string, err error) {
	path = root
	for _, name := range strings.Split(rel, "/") {
		next := filepath.Join(path, name)
		if _, statErr := os.Stat(next); statErr == nil {
			path = next
			continue
		}
		var entries []os.DirEntry
		if entries, err = os.ReadDir(path); err != nil {
			return "", err
		}
		found := false
		for _, e := range entries {
			if norm.NFC.String(e.Name()) == norm.NFC.String(name) {
				next, found = filepath.Join(path, e.Name()), true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("%s: %w", next, fs.ErrNotExist)
		}
		path = next
	}
	return path, nil
}
