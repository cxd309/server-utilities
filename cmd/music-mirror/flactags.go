package main

import (
	"os"
	"strings"

	"github.com/go-flac/flacvorbis"
	flac "github.com/go-flac/go-flac"
)

// readFLAC reads a FLAC file's tags
func readFLAC(path string) (tags trackTags, err error) {
	var v vorbisComments
	if v, err = readVorbisComments(path); err != nil {
		return trackTags{}, err
	}
	return fromVorbis(v), nil
}

// vorbisComments holds a FLAC file's comments, keyed by upper-case field
// name. A field can appear more than once, e.g. two GENRE entries.
type vorbisComments map[string][]string

// readVorbisComments reads only the FLAC header blocks, not the audio.
func readVorbisComments(path string) (v vorbisComments, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return nil, err
	}
	defer f.Close()

	var meta *flac.File
	if meta, err = flac.ParseMetadata(f); err != nil {
		return nil, err
	}

	v = vorbisComments{}
	for _, block := range meta.Meta {
		if block.Type != flac.VorbisComment {
			continue
		}
		var cmt *flacvorbis.MetaDataBlockVorbisComment
		if cmt, err = flacvorbis.ParseFromMetaDataBlock(*block); err != nil {
			return nil, err
		}
		for _, c := range cmt.Comments {
			key, val, ok := strings.Cut(c, "=")
			if !ok {
				continue // malformed entry; skip rather than fail the track
			}
			key = strings.ToUpper(key)
			v[key] = append(v[key], val)
		}
	}
	return v, nil
}

// get returns the values of the first listed field that is set, joined
// with "; ". Later fields are fallbacks.
func (v vorbisComments) get(fields ...string) (value string) {
	for _, field := range fields {
		if value = joinValues(v[field]); value != "" {
			return value
		}
	}
	return ""
}

// numberPair builds "n/total", e.g. TRACKNUMBER=3 and TRACKTOTAL=12
// give "3/12". The total is optional.
func (v vorbisComments) numberPair(numField string, totalFields ...string) (pair string) {
	n := v.get(numField)
	if n == "" || strings.Contains(n, "/") {
		return n
	}
	if total := v.get(totalFields...); total != "" {
		return n + "/" + total
	}
	return n
}

func fromVorbis(v vorbisComments) (tags trackTags) {
	return trackTags{
		Title:           v.get("TITLE"),
		Artist:          v.get("ARTIST"),
		Album:           v.get("ALBUM"),
		AlbumArtist:     v.get("ALBUMARTIST", "ALBUM ARTIST"),
		Genre:           v.get("GENRE"),
		Composer:        v.get("COMPOSER"),
		ArtistSort:      v.get("ARTISTSORT"),
		AlbumArtistSort: v.get("ALBUMARTISTSORT"),
		AlbumSort:       v.get("ALBUMSORT"),
		TitleSort:       v.get("TITLESORT"),
		Year:            yearOf(v.get("ORIGINALDATE"), v.get("DATE"), v.get("YEAR")),
		Track:           v.numberPair("TRACKNUMBER", "TRACKTOTAL", "TOTALTRACKS"),
		Disc:            v.numberPair("DISCNUMBER", "DISCTOTAL", "TOTALDISCS"),
		Compilation:     v.get("COMPILATION") == "1",
	}
}
