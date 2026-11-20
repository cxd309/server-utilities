package main

import (
	"os"
	"path/filepath"
	"strings"
)

// book is one book folder in the library, and where its parts go in the mirror
type book struct {
	src string // e.g. /Volumes/audiobooks/Terry Pratchett read by Tony Robinson/Mort
	dst string // the same Author/Title under the mirror root
}

// scanBooks finds every <Author>/<Title> folder under src.
// Audio files sitting directly in src or in an author folder aren't part of any book,
// so they're returned as stray for a warning.
// Hidden files and folders are skipped
func scanBooks(src, dest string) (books []book, stray []string, err error) {
	var authors []os.DirEntry
	if authors, err = os.ReadDir(src); err != nil {
		return nil, nil, err
	}
	for _, a := range authors {
		authorPath := filepath.Join(src, a.Name())
		if hidden(a.Name()) {
			continue
		}
		if !a.IsDir() {
			if isAudio(a.Name()) {
				stray = append(stray, authorPath)
			}
			continue
		}

		var titles []os.DirEntry
		if titles, err = os.ReadDir(authorPath); err != nil {
			return nil, nil, err
		}
		for _, t := range titles {
			titlePath := filepath.Join(authorPath, t.Name())
			if hidden(t.Name()) {
				continue
			}
			if !t.IsDir() {
				if isAudio(t.Name()) {
					stray = append(stray, titlePath)
				}
				continue
			}
			books = append(books, book{src: titlePath, dst: filepath.Join(dest, a.Name(), t.Name())})
		}
	}
	return books, stray, nil
}

// hidden reports whether name is a dotfile, such as .DS_Store or a build folder
func hidden(name string) (yes bool) {
	return strings.HasPrefix(name, ".")
}

// isAudio reports whether name is a file type the library holds books in
func isAudio(name string) (yes bool) {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".m4a", ".flac":
		return true
	}
	return false
}
