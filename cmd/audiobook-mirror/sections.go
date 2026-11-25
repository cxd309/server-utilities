package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// sectionsFile, in a book's folder, splits a collection into its sections, like the books
// in an omnibus. Each line is where a section starts, a space, then the section's name:
//
//	001 The Murder of Roger Ackroyd
//	2-16 The Million Dollar Bond Robbery
//
// The start is a chapter file's number, the part of its name before the first space.
// A section runs until the next one starts; chapters before the first have none.
// Blank lines and lines starting with # are skipped
const sectionsFile = "sections.txt"

// readSections returns the section of each chapter file in paths, from dir's sections.txt.
// Without one, every section is "".
// A line with no name, a start that isn't a chapter file, or a start that doesn't come
// after the one before is an error, so a typo can't quietly merge two sections
func readSections(dir string, paths []string) (sections []string, err error) {
	sections = make([]string, len(paths))
	var data []byte
	if data, err = os.ReadFile(filepath.Join(dir, sectionsFile)); errors.Is(err, fs.ErrNotExist) {
		return sections, nil
	} else if err != nil {
		return nil, err
	}

	starts := make(map[string]int, len(paths)) // a chapter file's number → its index in paths
	for i, path := range paths {
		number, _, _ := strings.Cut(filepath.Base(path), " ")
		starts[number] = i
	}

	previous := -1 // index where the last section started
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		number, name, _ := strings.Cut(line, " ")
		name = strings.TrimSpace(name)
		start, found := starts[number]
		switch {
		case name == "":
			return nil, &sectionLineError{Line: n + 1, Text: line, Err: errSectionNoName}
		case !found:
			return nil, &sectionLineError{Line: n + 1, Text: line, Err: errSectionStart}
		case start <= previous:
			return nil, &sectionLineError{Line: n + 1, Text: line, Err: errSectionOrder}
		}
		for i := start; i < len(paths); i++ {
			sections[i] = name
		}
		previous = start
	}
	return sections, nil
}
