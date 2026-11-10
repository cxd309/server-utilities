package main

import (
	"strings"
)

// trackTags is everything we put into a mirror MP3, whatever the source.
// Readers fill it in, writeID3 writes it out.
// Empty fields are skipped
type trackTags struct {
	Title, Artist, Album, AlbumArtist, Genre, Composer string
	ArtistSort, AlbumArtistSort, AlbumSort, TitleSort  string
	Year                                               string // four digits, original release where known
	Track, Disc                                        string // "n" or "n/total"
	Compilation                                        bool
}

// joinValues trims each value, drops empty ones and joins the rest with "; ".
func joinValues(values []string) (joined string) {
	var kept []string
	for _, s := range values {
		if s = strings.TrimSpace(s); s != "" {
			kept = append(kept, s)
		}
	}
	return strings.Join(kept, "; ")
}

// yearOf returns the four-digit year from the first date that has one,
// so callers list dates in order of preference.
func yearOf(dates ...string) (year string) {
	for _, d := range dates {
		if len(d) >= 4 && isDigits(d[:4]) {
			return d[:4]
		}
	}
	return ""
}

func isDigits(s string) (ok bool) {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
