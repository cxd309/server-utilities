package main

import (
	"context"

	"github.com/cxd309/server-utilities/internal/ffmpeg"
)

// coverName is what the cover is saved as in the book's folder.
// Picard's local files cover provider and audiobook-mirror both look for it
const coverName = "cover.jpg"

// coverArgs builds the ffmpeg command line that saves src's embedded picture to out as a JPEG.
// codec is the picture's codec: a JPEG ("mjpeg") is copied as it is,
// anything else (a PNG, say) is converted
func coverArgs(src, out, codec string) (args []string) {
	args = []string{"-i", src, "-map", "0:v:0", "-frames:v", "1"}
	if codec == "mjpeg" {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", "mjpeg", "-q:v", "2") // 2 is near the best JPEG quality
	}
	return append(args, "-update", "1", "-f", "image2", out) // one picture, not a numbered sequence
}

// saveCover saves src's embedded picture, whose codec is codec, to out
func saveCover(ctx context.Context, src, out, codec string) (err error) {
	return ffmpeg.Run(ctx, coverArgs(src, out, codec)...)
}
