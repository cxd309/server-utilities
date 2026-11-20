package main

import (
	"context"
	"errors"
	"os/exec"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// checkFFmpeg confirms ffmpeg is installed. The ALAC encoder is built into
// every ffmpeg, so unlike MP3 there's no optional library to check for.
func checkFFmpeg() (err error) {
	if _, err = exec.LookPath("ffmpeg"); errors.Is(err, exec.ErrNotFound) {
		return ffmpeg.ErrNotInstalled
	}
	return err
}

// metadataArgs turns tags into ffmpeg -metadata options. The ipod muxer
// maps these names onto iTunes atoms (track -> trkn, compilation -> cpil, ...).
func metadataArgs(tags trackTags) (args []string) {
	add := func(key, value string) {
		if value != "" {
			args = append(args, "-metadata", key+"="+value)
		}
	}
	add("title", tags.Title)
	add("artist", tags.Artist)
	add("album", tags.Album)
	add("album_artist", tags.AlbumArtist)
	add("genre", tags.Genre)
	add("composer", tags.Composer)
	add("date", tags.Year)
	add("track", tags.Track)
	add("disc", tags.Disc)
	if tags.Compilation {
		add("compilation", "1")
	}
	return args
}

// ffmpegArgs builds the command line that turns one FLAC into a tagged ALAC .m4a
func ffmpegArgs(src, out string, tags trackTags) (args []string) {
	args = []string{
		"-i", src,
		"-map", "0:a:0", // first audio stream only, which drops embedded cover art
		"-map_metadata", "-1", // copy no tags across; we write our own below
		// Older iPods only play ALAC up to 16-bit/48 kHz, so bring anything
		// bigger down to CD quality. A no-op for files that already are;
		// dither only kicks in when reducing bit depth.
		"-af", "aresample=osr=44100:osf=s16p:dither_method=triangular",
		"-c:a", "alac",
	}
	args = append(args, metadataArgs(tags)...)
	return append(args,
		"-f", "ipod", // out will end in .part, so name the format explicitly
		out,
	)
}

// encode converts src to a tagged ALAC file at out
func encode(ctx context.Context, src, out string, tags trackTags) (err error) {
	return ffmpeg.Run(ctx, ffmpegArgs(src, out, tags)...)
}
