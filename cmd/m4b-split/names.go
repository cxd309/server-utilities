package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// maxNameBytes keeps names well under the 255-byte limit most filesystems have,
// leaving room for the track number and extension
const maxNameBytes = 200

// unsafeChars can't appear in a file name somewhere these files travel:
// Linux, SMB share or macOS
const unsafeChars = `/\*?"<>|`

// safeName makes s usable as one file or folder name.
// Colons are dropped, so "Chapter 2: Gary" becomes "Chapter 2 Gary",
// other characters some filesystems reject become "_".
func safeName(s string) (name string) {
	name = strings.Map(func(r rune) rune {
		switch {
		case r == ':':
			return -1 // dropped
		case unicode.IsSpace(r):
			return ' ' // tabs and newlines become spaces, tidy collapses them
		case r < 0x20 || strings.ContainsRune(unsafeChars, r):
			return '_'
		}
		return r
	}, s)
	name = truncate(tidy(name), maxNameBytes)
	// A leading dot hides the file, trailing dots and spaces upset SMB
	if name = strings.Trim(name, ". "); name == "" {
		return "_"
	}
	return name
}

// truncate shortens s to at most n bytes without cutting a character in half
func truncate(s string, n int) (short string) {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// chapterFileName names chapter n of total, zero-padded so the files sort in order:
// "01 Opening Credits.m4a" ... "12 End Credits.m4a"
func chapterFileName(n, total int, title string) (name string) {
	width := max(2, len(strconv.Itoa(total)))
	return fmt.Sprintf("%0*d %s.m4a", width, n, safeName(title))
}
