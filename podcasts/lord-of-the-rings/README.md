# The Lord of the Rings (BBC Radio 4, 1981)

**Source:** the three CD sets of the BBC's 1981 dramatisation in the audiobook library, tagged with
Picard: `J.R.R. Tolkien/<The Fellowship of the Ring | The Two Towers | The Return of the King> (Audio Drama)`.
FLAC, one file per scene (32, 36 and 28 scenes), cover in each folder. These sets dropped the
26 broadcast episodes' divisions, so scenes are the only breaks left.

**Result:** three shows, one per book, one episode per scene, made with `podcast-import` from
`episodes.csv` (all three shows are in the one manifest). FLAC becomes ALAC.

## Decisions

- **Show names** are numbered, "The Lord of the Rings 1 - The Fellowship of the Ring" and so on,
  so the iPod lists the books together and in order (alphabetically the Return comes before the Two).
  No colons, which a Windows share can't show.
- **Titles** are the scene titles from the tags. One MusicBrainz typo is corrected here:
  "The Steward of Minas Tirlith" is titled "The Steward of Minas Tirith".
- **Release dates:** scenes have none, so each book starts on the first broadcast of the episode
  where the book began (episodes 1, 10 and 18: 8 March, 10 May and 5 July 1981, from Wikipedia's
  page on the series) and the scenes follow an hour apart, purely to keep them in order.
  The description of each gives its scene, disc and track.
- **Left out:** The Return of the King's disc 4 (Stephen Oliver's music, 19 tracks) and its
  "End Credits" track.

## How it was run

```sh
go run ./cmd/podcast-import --dry-run podcasts/lord-of-the-rings/episodes.csv
go run ./cmd/podcast-import podcasts/lord-of-the-rings/episodes.csv
```
