package main

import "errors"

var (
	errNoAudio      = errors.New("file has no audio stream")
	errNoChapters   = errors.New("file has no chapters to split on")
	errAlreadySplit = errors.New("already split")
	errPieceCount   = errors.New("ffmpeg made a different number of pieces than there are chapters")
)
