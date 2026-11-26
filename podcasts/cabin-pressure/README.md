# Cabin Pressure

**Source:** _The Complete Cabin Pressure: From A to Z_ (BBC, 2015), 14 CDs, ripped with
ripcd and tagged with Picard against MusicBrainz release `819af8c6-69fc-4a0e-95d1-b168ccc75c52`.
Library folder:
`/mnt/prudence/audiobooks/John Finnemore/The Complete Cabin Pressure From A to Z (Audio Drama)`.
Each episode is split into CD tracks named "Abu Dhabi, TR 1" … "TR 9" (the interview has 12),
two episodes per disc.

**Result:** 28 files in `/mnt/prudence/podcasts/Cabin Pressure/`, one per episode:
lossless ALAC (16-bit/44.1 kHz) with the box set cover, tagged as a podcast
(show "Cabin Pressure", artist John Finnemore, media kind Podcast).

## Decisions

- **Order: A-Z**, reading online it seems this is the order they were written hence the boxset
- **Release dates** use real release dates but alphabetical order, some s02&03 episodes are swapped
- **The interview** on disc 14 is dated to the box set's release, 5 February 2015.
- Episode data is from the BBC programme pages (their `.json` versions give
  `first_broadcast_date`), collected into `episodes.csv`, the CSV is built into the binary.

## How it works

A one-off program, `one-off/cabin-pressure/main.go` (deleted since; it's in git history),
grouped the CD tracks by the name before the last ", TR ", paired each group with the CSV row
in the same position (stopping if the CD name wasn't in the CSV title), joined the tracks with
ffmpeg's concat demuxer into ALAC, wrote the tags, and then set the podcast flag with `internal/podcast` (Python's mutagen): iTunes silently refuses a file whose media
kind is Podcast without it, and ffmpeg can't write it. Existing episodes were
skipped, so it could be re-run.

`podcast-import` can't do this one: it makes one episode per file, and these episodes were
each about nine CD tracks joined together.

```sh
go run ./one-off/cabin-pressure --dry-run   # the plan, as it was run
go run ./one-off/cabin-pressure             # writes the episodes
```

**Source removed:** once the episodes were in `/mnt/prudence/podcasts`, the source folder was
deleted from the audiobook library (2026-11-25), so this script can no longer be re-run as is.
