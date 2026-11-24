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

`main.go` groups the CD tracks by the name before the last ", TR ", pairs each group with
the CSV row in the same position (stopping if the CD name isn't in the CSV title), joins the
tracks with ffmpeg's concat demuxer into ALAC, writes the tags, and then sets the podcast
flag with `internal/podcast` (Python's mutagen): iTunes silently refuses a file whose media
kind is Podcast without it, and ffmpeg can't write it. Existing episodes are
skipped, so it can be re-run.

```sh
go run ./one-off/cabin-pressure --dry-run   # the plan
go run ./one-off/cabin-pressure             # writes the episodes
```
