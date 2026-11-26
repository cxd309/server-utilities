package main

// Limits for one audiobook part. A part stops at whichever it reaches first.
// Tested on the 7th gen iPod classic in November 2026: every part over 13.5 h lost its chapter
// markers, and skipping to 15 h in a 19 h part crashed it, while parts up to 2.7 GB were fine.
// To change them, edit and rebuild.
const (
	// samplesPerFrame is the size of one AAC-LC frame
	samplesPerFrame = 1024

	// maxPartSamples is the most a signed 32-bit count of samples can hold,
	// which is how the iPod reads an MP4's duration (the audio track's time scale is its
	// sample rate). It's MarkAble's 13.5 h at 44.1 kHz. At 22 kHz it allows 27 h, which is untested
	maxPartSamples = 1<<31 - 1

	// maxPartBytes keeps a part under 4 GB, where an MP4 needs 64-bit file offsets.
	// The 100 MB to spare covers the cover, chapter track and headers.
	// MarkAble's 800 MB turned out to be needlessly cautious
	maxPartBytes = 3900 * 1000 * 1000

	// alacOverFLAC is how much bigger ALAC is than the FLAC it's made from.
	// Both are lossless, but ALAC compresses slightly less well. 10% is a safe margin
	alacOverFLAC = 1.1

	// pcmBytesPerFrame is one frame of uncompressed 16-bit stereo,
	// the most ALAC can take when it's made from AAC
	pcmBytesPerFrame = samplesPerFrame * 2 * 2
)

// For the iPod mini: it skipped an AAC part of 1,125,350 frames but played one of 1,022,745,
// so it seems to stop at 2²⁰ frames (about 6.6 h at 44.1 kHz). Set maxPartSamples to
// 1<<20 * samplesPerFrame and rebuild. It also scratches on 22 kHz AAC.
