package main

import (
	"os"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacvorbis"
	flac "github.com/go-flac/go-flac"
)

// vorbisTags holds a FLAC file's Vorbis comments, keyed by upper-case field name.
// A field can appear more than once, e.g. two GENRE entries.
type vorbisTags map[string][]string

// readFLACTags reads the Vorbis comments from a FLAC file's header
func readFLACTags(path string) (tags vorbisTags, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return nil, err
	}
	defer f.Close()

	var meta *flac.File
	if meta, err = flac.ParseMetadata(f); err != nil {
		return nil, err
	}

	tags = vorbisTags{}
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
			tags[key] = append(tags[key], val)
		}
	}
	return tags, nil
}

// get returns the values of the first listed field that is set, joined with "; ".
// Fields are tried in order, so later ones are fallbacks.
func (t vorbisTags) get(fields ...string) (value string) {
	for _, field := range fields {
		var vals []string
		for _, v := range t[field] {
			if v = strings.TrimSpace(v); v != "" {
				vals = append(vals, v)
			}
		}
		if len(vals) > 0 {
			return strings.Join(vals, "; ")
		}
	}
	return ""
}

// numberPair builds an ID3 "n/total" value, e.g. TRACKNUMBER=3 and TRACKTOTAL=12 give "3/12"
// The total is optional
func (t vorbisTags) numberPair(numField string, totalFields ...string) (pair string) {
	n := t.get(numField)
	if n == "" || strings.Contains(n, "/") {
		return n
	}
	if total := t.get(totalFields...); total != "" {
		return n + "/" + total
	}
	return n
}

// year returns the four-digit year from DATE Picard writes e.g. 1999-03-04
func (t vorbisTags) year() (year string) {
	d := t.get("DATE", "ORIGINALDATE", "YEAR")
	if len(d) < 4 {
		return ""
	}
	for _, r := range d[:4] {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return d[:4]
}

// id3TextFrames maps ID3 frames to the Vorbis fields that feed them.
var id3TextFrames = []struct {
	frame  string
	fields []string
}{
	{"TIT2", []string{"TITLE"}},
	{"TPE1", []string{"ARTIST"}},
	{"TALB", []string{"ALBUM"}},
	{"TPE2", []string{"ALBUMARTIST", "ALBUM ARTIST"}}, // album artist
	{"TCON", []string{"GENRE"}},
	{"TCOM", []string{"COMPOSER"}},
	{"TSOP", []string{"ARTISTSORT"}},      // sort artist
	{"TSO2", []string{"ALBUMARTISTSORT"}}, // sort album artist (iTunes frame)
	{"TSOA", []string{"ALBUMSORT"}},       // sort album
	{"TSOT", []string{"TITLESORT"}},       // sort title
}

// writeID3 writes tags into an untagged MP3 as ID3v2.3
func writeID3(mp3Path string, tags vorbisTags) (err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(mp3Path, id3v2.Options{Parse: false}); err != nil {
		return err
	}
	defer tag.Close()

	tag.SetVersion(3)
	set := func(frame, text string) {
		if text != "" {
			tag.AddTextFrame(frame, id3v2.EncodingUTF16, text)
		}
	}

	for _, m := range id3TextFrames {
		set(m.frame, tags.get(m.fields...))
	}
	set("TYER", tags.year())
	set("TRCK", tags.numberPair("TRACKNUMBER", "TRACKTOTAL", "TOTALTRACKS"))
	set("TPOS", tags.numberPair("DISCNUMBER", "DISCTOTAL", "TOTALDISCS"))
	if tags.get("COMPILATION") == "1" {
		set("TCMP", "1") // iTunes compilation flag: groups "Various Artists" albums
	}
	return tag.Save()
}
