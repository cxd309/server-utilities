# Bleak Expectations

**Source:** the Audible download of *Bleak Expectations: The Complete BBC Radio 4 Series*,
split with m4b-split and tagged with Picard. Library folder:
`/mnt/prudence/audiobooks/Mark Evans/Bleak Expectations The Complete BBC Radio 4 Series (Audio Drama)`.
30 chapters, one per episode, titled only "Chapter 1" … "Chapter 30". AAC, 44.1 kHz.
No embedded artwork.

**Result:** 30 files in `/mnt/prudence/podcasts/Bleak Expectations/`, one per episode, with
the audio copied unchanged (no re-encoding), tagged as a podcast (show "Bleak Expectations",
artist Mark Evans, media kind Podcast plus the podcast flag).

## Decisions

- **Titles** are the broadcast titles ("A Childhood Cruelly Kippered"), with no series number.
- **Release dates** are the real first Radio 4 broadcasts, which are already in order
  (5 series of 6 episodes, 2007–2012). The series and episode are in each description.
- Episode data is from the BBC programme pages, collected into `episodes.csv`.
- **Artwork:** the download has none. Put a `cover.jpg` in this folder before running and it's
  embedded in every episode; without one the episodes have no artwork.

## How it works

`main.go` checks the files are Chapter 1 … 30 in order, pairs each with the CSV row in the
same position, copies the audio with ffmpeg while writing the tags, and sets the podcast flag
with `internal/podcast`. Existing episodes are skipped, so it can be re-run.

```sh
go run ./one-off/bleak-expectations --dry-run
go run ./one-off/bleak-expectations
```
