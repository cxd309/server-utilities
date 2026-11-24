package main

import (
	"strings"
)

// trackTags is everything we put into a mirror MP3, whatever the source.
// Readers fill it in, writeID3 writes it out.
// Empty fields are skipped
type trackTags struct {
	Title, Artist, Album, AlbumArtist, Genre, Composer string
	Year                                               string // four digits, original release where known
	Track, Disc                                        string // "n" or "n/total"
	Compilation                                        bool
	Cover                                              *cover // embedded art, MP3 sources only; nil for none
}

// cover is an embedded picture: the front cover, as Picard embeds it.
// FLAC covers go straight through ffmpeg, so only MP3s carry one here
type cover struct {
	mime string // e.g. "image/jpeg"
	data []byte
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

// splitCredit separates an artist credit into its primary artist and the rest,
// using the credited artists Picard stores in ARTISTS.
// So "Calvin Harris feat. Rihanna" with ["Calvin Harris", "Rihanna"]
// gives "Calvin Harris" and "feat. Rihanna".
// A credit with one artist is returned whole, one that doesn't start with its first artist loses the rest.
func splitCredit(credit string, artists []string) (primary, rest string) {
	if len(artists) < 2 {
		return credit, ""
	}
	primary = artists[0]
	if !strings.HasPrefix(credit, primary) {
		return primary, ""
	}
	return primary, strings.TrimSpace(strings.TrimPrefix(credit, primary))
}

// titleWithGuests appends the rest of a credit to the title, Apple style:
// "feat. Rihanna" gives "Title (feat. Rihanna)", and any other join
// ("& David Bowie", "x Stormzy", ", B & C") gives "Title (with …)".
// Titles that already name a featured artist are left alone.
func titleWithGuests(title, rest string) (full string) {
	join, guests, ok := strings.Cut(rest, " ")
	if !ok || guests == "" || strings.Contains(strings.ToLower(title), "feat.") {
		return title
	}
	switch strings.ToLower(join) {
	case "feat.", "ft.", "featuring":
		return title + " (feat. " + guests + ")"
	}
	return title + " (with " + guests + ")"
}
