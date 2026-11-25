package main

import (
	"fmt"
	"strconv"
	"strings"
)

// bookTags is what every part of a book is tagged with
type bookTags struct {
	title    string // the album: the album tag, until buildBook swaps in the folder name, e.g. "Mort (Tony Robinson)"
	author   string // the MusicBrainz credit, e.g. "Terry Pratchett read by Tony Robinson"
	narrator string // may be empty
	date     string // may be empty
}

// partTitles names each part of a book, given the section of each part in order.
// Parts of a named section are named after it: "Animal Farm", or "1984, Part 1 of 2"
// when the section needed more than one. Other parts are named after the book:
// "Mort", or "Mort, Part 02 of 12" counted across the whole book
func partTitles(book string, sections []string) (titles []string) {
	total := len(sections)
	for i, s := range sections {
		if s == "" {
			titles = append(titles, numbered(book, i+1, total))
			continue
		}
		first, last := i, i
		for first > 0 && sections[first-1] == s {
			first--
		}
		for last < total-1 && sections[last+1] == s {
			last++
		}
		titles = append(titles, numbered(s, i-first+1, last-first+1))
	}
	return titles
}

// numbered is name on its own when there's one part,
// otherwise "name, Part n of total", padded so 12 parts give "Part 02 of 12"
func numbered(name string, n, total int) (title string) {
	if total <= 1 {
		return name
	}
	title = fmt.Sprintf("%s, Part %0*d of %d", name, digits(total), n, total)
	return title
}

// digits is how many digits n has
func digits(n int) (count int) {
	count = len(strconv.Itoa(n))
	return count
}

// tagLines is the global tags for part n of total, in ffmetadata format, titled title.
// The track number keeps the parts in order. (A sort title would too, but ffmpeg's
// iPod writer drops sort_name; if the iPod sorts parts by title, number the titles instead.)
// Empty values are left out
func tagLines(book bookTags, title string, n, total int) (lines string) {
	var b strings.Builder
	tag := func(key, value string) {
		if value != "" {
			fmt.Fprintf(&b, "%s=%s\n", key, metaEscape(value))
		}
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
