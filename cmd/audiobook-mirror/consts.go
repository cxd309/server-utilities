package main

// Limits for one audiobook part. A part stops at whichever it reaches first.
// They're the limits of an MP4 file with 32-bit headers, which the iPod classic should read.
// (MarkAble's defaults, 13.5 h and 800 MB, were safer guesses that split books more.)
// To change them, edit and rebuild.
const (
	// samplesPerFrame is the size of one AAC-LC frame
	samplesPerFrame = 1024

	// maxPartSamples is the most an MP4's 32-bit duration fields can hold,
	// counted in samples (the audio track's time scale is its sample rate):
	// about 27 h at 44.1 kHz, or 54 h at 22 kHz.
	// Longer files need 64-bit headers, which iPod firmware may not read.
	// MarkAble took the fields as signed, half this: 13.5 h at 44.1 kHz
	maxPartSamples = 1<<32 - 1

	// maxPartBytes keeps a part under 4 GB, where an MP4 needs 64-bit file offsets.
	// The 100 MB to spare covers the cover, chapter track and headers
	maxPartBytes = 3900 * 1000 * 1000

	// alacOverFLAC is how much bigger ALAC is than the FLAC it's made from.
	// Both are lossless, ALAC compresses slightly less well; 10% is a safe margin
	alacOverFLAC = 1.1

	// pcmBytesPerFrame is one frame of uncompressed 16-bit stereo,
	// the most ALAC can take when it's made from AAC
	pcmBytesPerFrame = samplesPerFrame * 2 * 2
)

// For the iPod mini: it skipped an AAC part of 1,125,350 frames but played one of 1,022,745,
// so it seems to stop at 2²⁰ frames (about 6.6 h at 44.1 kHz). Set maxPartSamples to
// 1<<20 * samplesPerFrame and rebuild. It also scratches on 22 kHz AAC.
