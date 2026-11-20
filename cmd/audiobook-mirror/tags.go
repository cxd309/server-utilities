package main

import (
	"fmt"
	"strconv"
	"strings"
)

// bookTags is what every part of a book is tagged with
type bookTags struct {
	title    string // e.g. "Mort"
	author   string // the MusicBrainz credit, e.g. "Terry Pratchett read by Tony Robinson"
	narrator string // may be empty
	date     string // may be empty
}

// tagLines is the global tags for part n of total, in ffmetadata format.
// Empty values are left out
func tagLines(book bookTags, n, total int) (lines string) {
	var b strings.Builder
	tag := func(key, value string) {
		if value != "" {
			fmt.Fprintf(&b, "%s=%s\n", key, metaEscape(value))
		}
	}

	title := book.title
	if total > 1 {
		width := len(strconv.Itoa(total)) // 7 parts: "Part 1 of 7"; 12 parts: "Part 01 of 12"
		title = fmt.Sprintf("%s, Part %0*d of %d", book.title, width, n, total)
	}

	tag("title", title)
	tag("album", book.title)
	tag("artist", book.author)
	tag("album_artist", book.author)
	tag("composer", book.narrator)
	tag("date", book.date)
	tag("track", fmt.Sprintf("%d/%d", n, total))
	tag("genre", "Audiobook")
	tag("media_type", "2") // iTunes "Audiobook"
	lines = b.String()
	return lines
}
