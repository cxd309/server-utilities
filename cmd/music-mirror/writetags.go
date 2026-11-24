package main

import "github.com/bogem/id3v2/v2"

// writeID3 gives the MP3 at path a fresh ID3v2.3 tag containing only tags.
// Any existing tag is replaced; cover art is only kept if tags.Cover carries it.
func writeID3(path string, tags trackTags) (err error) {
	var tag *id3v2.Tag
	// Parse: false still notes the size of any existing tag,
	// so Save replaces it, but none of its frames are loaded or kept
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: false}); err != nil {
		return err
	}
	defer tag.Close()

	tag.SetVersion(3)
	set := func(id, text string) {
		if text != "" {
			// v2.3 has no UTF-8, without this the library defaults to Latin-1
			tag.AddTextFrame(id, id3v2.EncodingUTF16, text)
		}
	}

	set("TIT2", tags.Title)
	set("TPE1", tags.Artist)
	set("TALB", tags.Album)
	set("TPE2", tags.AlbumArtist)
	set("TCON", tags.Genre)
	set("TCOM", tags.Composer)
	set("TYER", tags.Year) // v2.3 year frame (v2.4 calls it TDRC)
	set("TRCK", tags.Track)
	set("TPOS", tags.Disc)
	if tags.Compilation {
		set("TCMP", "1") // iTunes compilation flag
	}
	if tags.Cover != nil {
		tag.AddAttachedPicture(id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF16,
			MimeType:    tags.Cover.mime,
			PictureType: id3v2.PTFrontCover,
			Picture:     tags.Cover.data,
		})
	}
	return tag.Save()
}
