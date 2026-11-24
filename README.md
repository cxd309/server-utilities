# server-utilities

Tools for my music and audiobook library, and for getting it onto an iPod.

- `ripcd` - rip CDs with whipper, matched against MusicBrainz with a given barcode
- `music-mirror` - mirror the music library for the iPod: FLAC becomes ALAC, MP3 is copied and re-tagged
- `music-sync` - keep the Music app's library in step with the music mirror
- `m4b-split` - split an Audible `.m4b` into one file per chapter, losslessly, ready for Picard
- `audiobook-mirror` - join each audiobook's chapters into iPod-safe `.m4b` parts with chapter markers

## Picard

Both libraries are tagged with [MusicBrainz Picard](https://picard.musicbrainz.org/).
Its file naming scripts decide the folder layout the tools above expect,
so they're kept here, in `picard/`, and changed in git like any other code.

| Script                   | Layout                                                                            |
| ------------------------ | --------------------------------------------------------------------------------- |
| `picard/music.ptsp`      | `Album artist/Album/[disc-]NN Title`, into `/mnt/prudence/music`                  |
| `picard/audiobooks.ptsp` | `Author/Book (Narrator)/[disc-]NN Chapter title`, into `/mnt/prudence/audiobooks` |

They're Picard's own script export format: plain text, with the script in an indented block.

To use one, open Picard's file naming script editor (Options → File Naming) and **Import** it.
After changing a script in Picard, **Export** it over the file here and commit,
then re-save the affected albums so their files move.
The destination directory is a separate Picard setting; each script notes the one it expects.

## One-off jobs

`one-off/` holds scripts for unusual releases, run once by hand and kept as worked examples.
See `one-off/README.md`.
