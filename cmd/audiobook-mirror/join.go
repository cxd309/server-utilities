package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// chapterFile is one chapter file of a downloaded book
type chapterFile struct {
	path    string
	title   string
	frames  int64
	codec   string // ffprobe's codec name, such as "aac" or "flac"
	section string // from sections.txt, e.g. "Animal Farm"; "" for none
	bytes   int64  // file size, for estimating how big parts will be
}

// frameTime is the time where frame starts, as ffmpeg seconds with six decimals.
// It rounds down to the microsecond,
// because a cut that's late by even a little makes ffmpeg copy the frame into both parts
func frameTime(frame int64, sampleRate int) (s string) {
	us := frame * samplesPerFrame * 1_000_000 / int64(sampleRate)
	s = fmt.Sprintf("%d.%06d", us/1_000_000, us%1_000_000)
	return s
}

// concatQuote wraps path in single quotes for an ffmpeg concat list.
// A quote inside the path is written as '\” : close the quotes, an escaped quote, open them again
func concatQuote(path string) (quoted string) {
	quoted = "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
	return quoted
}

// concatList is the ffmpeg concat list for one part.
// A piece that starts or ends partway through its chapter gets an inpoint or outpoint
// With preroll, a piece starting mid-chapter starts one frame early (see joinArgs)
func concatList(part []piece, chapters []chapterFile, sampleRate int, preroll bool) (list string) {
	var b strings.Builder
	for _, p := range part {
		c := chapters[p.chapter]
		fmt.Fprintf(&b, "file %s\n", concatQuote(c.path))
		if p.start > 0 {
			start := p.start
			if preroll {
				start-- // the decoder needs the frame before to get the first one right
			}
			fmt.Fprintf(&b, "inpoint %s\n", frameTime(start, sampleRate))
		}
		if p.end < c.frames {
			fmt.Fprintf(&b, "outpoint %s\n", frameTime(p.end, sampleRate))
		}
	}
	list = b.String()
	return list
}

// metaEscaper escapes the characters that mean something in ffmpeg's metadata format
var metaEscaper = strings.NewReplacer(`\`, `\\`, "=", `\=`, ";", `\;`, "#", `\#`, "\n", "\\\n")

// metaEscape makes s safe to use as a value in an ffmetadata file
func metaEscape(s string) (escaped string) {
	escaped = metaEscaper.Replace(s)
	return escaped
}

// chapterMeta is an ffmetadata file with one chapter marker per piece of the part.
// Times are sample counts (TIMEBASE 1/sampleRate), counted from the start of the part
func chapterMeta(part []piece, chapters []chapterFile, sampleRate int) (meta string) {
	var b strings.Builder

	var pos int64 // frames into the part where this piece starts
	for _, p := range part {
		size := p.end - p.start
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/%d\nSTART=%d\nEND=%d\ntitle=%s\n",
			sampleRate,
			pos*samplesPerFrame,
			(pos+size)*samplesPerFrame,
			metaEscape(chapters[p.chapter].title))
		pos += size
	}
	meta = b.String()
	return meta
}

// joinArgs is the ffmpeg command line that joins the pieces in the concat list
// into one part at out, with the tags and chapter markers from the metadata file.
// cover is the path of the cover image, or "" for none.
// With copyAudio the AAC is copied as it is; otherwise it's decoded and stored as 16-bit ALAC
// at its own sample rate. preroll drops the extra frame concatList adds before a mid-chapter start
func joinArgs(listPath, metaPath, cover, out string, copyAudio, preroll bool) (args []string) {
	args = []string{
		"-f", "concat",
		"-safe", "0", // the list holds absolute paths
		"-i", listPath,
		"-i", metaPath,
	}
	if cover != "" {
		args = append(args, "-i", cover)
	}

	args = append(args, "-map", "0:a:0")
	if cover != "" {
		args = append(args, "-map", "2:v:0", "-disposition:v", "attached_pic") // one picture, even from a chapter file
	}

	if preroll {
		args = append(args, "-af", fmt.Sprintf("atrim=start_sample=%d,asetpts=PTS-STARTPTS", samplesPerFrame))
	}

	args = append(args, "-map_metadata", "1", "-map_chapters", "1")
	if copyAudio {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "alac", "-sample_fmt", "s16p") // 16-bit, the most the iPod's ALAC decoder takes
	}
	return append(args,
		"-c:v", "copy", // the cover is already a JPEG
		"-f", "ipod",
		out,
	)
}

// joinPart joins the pieces of one part into the file at out.
// tags are the part's global tags from tagLines, and cover is as for joinArgs.
// The concat list and metadata file go in a temporary folder,
// which is removed afterwards
func joinPart(ctx context.Context, part []piece, chapters []chapterFile, sampleRate int, tags, cover, out string) (err error) {
	var dir string
	if dir, err = os.MkdirTemp("", "audiobook-mirror-"); err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	listPath := filepath.Join(dir, "list.txt")
	metaPath := filepath.Join(dir, "meta.txt")
	meta := ";FFMETADATA1\n" + tags + chapterMeta(part, chapters, sampleRate)
	// A book never mixes codecs, so the part's first chapter speaks for the rest
	copyAudio := canCopy(chapters[part[0].chapter].codec)
	preroll := !copyAudio && part[0].start > 0 // only a part's first piece can start mid-chapter
	if err = os.WriteFile(listPath, []byte(concatList(part, chapters, sampleRate, preroll)), 0o600); err != nil {
		return err
	}
	if err = os.WriteFile(metaPath, []byte(meta), 0o600); err != nil {
		return err
	}
	return ffmpeg.Run(ctx, joinArgs(listPath, metaPath, cover, out, copyAudio, preroll)...)
}

// canCopy reports whether a book's audio can go into its parts unchanged.
// The iPod plays AAC at any sample rate (22 kHz scratched on the Mac-formatted mini,
// to be retested on the Windows-formatted iPods) but can't play FLAC,
// so anything other than AAC is converted to ALAC
func canCopy(codec string) (yes bool) {
	yes = codec == "aac"
	return yes
}

// savePicture saves the picture embedded in src, a chapter file, as cover.jpg in a new
// temporary folder, and returns its path. The caller removes the folder
func savePicture(ctx context.Context, src string) (path string, err error) {
	var dir string
	if dir, err = os.MkdirTemp("", "audiobook-mirror-cover-"); err != nil {
		return "", err
	}
	path = filepath.Join(dir, "cover.jpg")
	if err = ffmpeg.Run(ctx, "-i", src, "-map", "0:v:0", "-frames:v", "1", "-c:v", "copy", "-f", "image2", path); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return path, nil
}
