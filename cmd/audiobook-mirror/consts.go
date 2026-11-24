package main

// Limits for one audiobook part. A part stops at whichever it reaches first.
// They follow MarkAble's defaults (thegriggs.org/david/software/markable),
// which were made for iPods of the classic's era. To change them, edit and rebuild.
const (
	// samplesPerFrame is the size of one AAC-LC frame
	samplesPerFrame = 1024

	// maxPartSeconds is MarkAble's default maximum length: 13.5 hours
	maxPartSeconds = 13*60*60 + 30*60

	// maxPartSamples is the most an MP4's 32-bit duration fields can hold,
	// counted in samples (the audio track's time scale is its sample rate).
	// Longer files need 64-bit headers, which older iPod firmware may not read.
	// It's 13.5 h at 44.1 kHz, but 27 h at 22 kHz, so maxPartSeconds still applies there
	maxPartSamples = 1<<31 - 1

	// maxPartBytes is MarkAble's default maximum file size: 800 MB.
	// At Audible's ~125 kbps, 13.5 h is about 760 MB, so this mostly limits ALAC parts
	maxPartBytes = 800 * 1000 * 1000

	// alacOverFLAC is how much bigger ALAC is than the FLAC it's made from.
	// Both are lossless, ALAC compresses slightly less well; 10% is a safe margin
	alacOverFLAC = 1.1

	// pcmBytesPerFrame is one frame of uncompressed 16-bit stereo,
	// the most ALAC can take when it's made from AAC
	pcmBytesPerFrame = samplesPerFrame * 2 * 2
)

// For the iPod mini, which crashed on files above ~300k AAC frames (~2 h at 44.1 kHz),
// set maxPartSeconds to 2*60*60 and rebuild.
