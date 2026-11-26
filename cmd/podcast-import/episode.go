package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
	"github.com/cxd309/server-utilities/internal/podcast"
)

// fileName is the file for episode n of a show, e.g. "01 Fit the First.m4a"
func fileName(n int, title string) (name string) {
	name = fmt.Sprintf("%02d %s.m4a", n, strings.NewReplacer("/", "-", ":", "").Replace(title))
	return name
}

// source is what podcast-import needs to know about a source file
type source struct {
	path  string
	codec string // the audio codec, e.g. "aac" or "flac"
	cover string // the cover's path: path itself when it's embedded, a cover.jpg beside it, or ""
}

// copyArgs makes episode n of total from src at out, tagged as a podcast.
// AAC is copied as it is; anything else, such as FLAC from a CD, becomes 16-bit ALAC
func copyArgs(src source, out string, ep episode, n, total int) (args []string) {
	args = []string{"-i", src.path}
	switch src.cover {
	case "":
		args = append(args, "-map", "0:a:0")
	case src.path:
		args = append(args, "-map", "0:a:0", "-map", "0:v:0")
	default:
		args = append(args, "-i", src.cover, "-map", "0:a:0", "-map", "1:v:0")
	}
	args = append(args, "-map_metadata", "-1", "-map_chapters", "-1")
	if src.codec == "aac" {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "alac", "-sample_fmt", "s16p") // 16-bit, the most the iPod's ALAC decoder takes
	}
	if src.cover != "" {
		args = append(args, "-c:v", "copy", "-disposition:v", "attached_pic")
	}
	return append(args,
		"-metadata", "title="+ep.title,
		"-metadata", "album="+ep.show,
		"-metadata", "artist="+ep.artist,
		"-metadata", "album_artist="+ep.artist,
		"-metadata", "date="+ep.releaseDate,
		"-metadata", fmt.Sprintf("track=%d/%d", n, total),
		"-metadata", "genre=Podcast",
		"-metadata", "description="+ep.description,
		"-metadata", "media_type=21", // iTunes "Podcast"
		"-f", "ipod", out,
	)
}

// makeEpisode writes one episode via a .part file, renamed into place when complete
func makeEpisode(ctx context.Context, src source, out string, ep episode, n, total int) (err error) {
	part := out + ".part"
	defer func() {
		if err != nil {
			_ = os.Remove(part)
		}
	}()
	if err = ffmpeg.Run(ctx, copyArgs(src, part, ep, n, total)...); err != nil {
		return err
	}
	// Without this flag iTunes silently refuses the file, even though its media kind is Podcast
	if err = podcast.Flag(ctx, part); err != nil {
		return err
	}
	return os.Rename(part, out)
}

// probeSource finds a source file's codec and cover
func probeSource(ctx context.Context, path string) (src source, err error) {
	var info ffmpeg.Info
	if info, err = ffmpeg.Probe(ctx, path); err != nil {
		return source{}, err
	}
	src.path = path
	for _, s := range info.Streams {
		switch {
		case s.CodecType == "audio" && src.codec == "":
			src.codec = s.CodecName
		case s.CodecType == "video" && src.cover == "": // ffprobe lists an embedded picture as video
			src.cover = path
		}
	}
	if src.codec == "" {
		return source{}, fmt.Errorf("%s: %w", path, errNoAudio)
	}
	if src.cover == "" {
		beside := filepath.Join(filepath.Dir(path), "cover.jpg")
		if _, statErr := os.Stat(beside); statErr == nil {
			src.cover = beside
		}
	}
	return src, nil
}
