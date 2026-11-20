package main

import (
	"strings"

	"github.com/bogem/id3v2/v2"
)

// readMP3 reads an MP3 file's ID3v2 tag (v2.3 or v2.4)
func readMP3(path string) (tags trackTags, err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true}); err != nil {
		return trackTags{}, err
	}
	defer tag.Close()
	return fromID3(tag), nil
}

// userText returns the values of the TXXX frame with the given description,
// e.g. Picard's "ARTISTS". v2.4 separates multiple values with NUL bytes.
func userText(tag *id3v2.Tag, description string) (values []string) {
	for _, f := range tag.GetFrames(tag.CommonID("User defined text information frame")) {
		frame, ok := f.(id3v2.UserDefinedTextFrame)
		if !ok || frame.Description != description {
			continue
		}
		for _, v := range strings.Split(frame.Value, "\x00") {
			if v = strings.TrimSpace(v); v != "" {
				values = append(values, v)
			}
		}
		return values
	}
	return nil
}

func fromID3(tag *id3v2.Tag) (tags trackTags) {
	// text returns the first listed frame that is set. v2.4 separates
	// multiple values with NUL bytes, so split and re-join them.
	text := func(ids ...string) (value string) {
		for _, id := range ids {
			if value = joinValues(strings.Split(tag.GetTextFrame(id).Text, "\x00")); value != "" {
				return value
			}
		}
		return ""
	}
	artist, guests := splitCredit(text("TPE1"), userText(tag, "ARTISTS"))

	return trackTags{
		Title:           titleWithGuests(text("TIT2"), guests),
		Artist:          artist,
		Album:           text("TALB"),
		AlbumArtist:     text("TPE2"),
		Genre:           text("TCON"),
		Composer:        text("TCOM"),
		ArtistSort:      text("TSOP"),
		AlbumArtistSort: text("TSO2"),
		AlbumSort:       text("TSOA"),
		TitleSort:       text("TSOT"),
		// Original release (v2.4 TDOR, v2.3 TORY), then recording/release date.
		Year:        yearOf(text("TDOR", "TORY"), text("TDRC", "TYER")),
		Track:       text("TRCK"),
		Disc:        text("TPOS"),
		Compilation: text("TCMP") == "1",
	}
}
