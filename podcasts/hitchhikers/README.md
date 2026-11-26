# The Hitchhiker's Guide to the Galaxy

**Source:** the five radio phases in the audiobook library, CD releases tagged with Picard:
`Douglas Adams/The Hitchhiker’s Guide to the Galaxy <Primary … Quintessential> Phase (Audio Drama)`.
One AAC file per Fit (episode), 44.1 kHz, cover embedded.

**Result:** the 26 Fits and two extras as one 28-episode show,
`/mnt/prudence/podcasts/The Hitchhiker’s Guide to the Galaxy/`, audio copied unchanged, made with `podcast-import` from `episodes.csv`.

## Decisions

- **One show** for all five phases, episodes titled "Fit the First" … "Fit the Twenty-Sixth".
  The phase is in each description, and each episode keeps its own phase's cover.
- **Release dates** are the first BBC Radio 4 broadcasts (1978–2005), from Wikipedia's Primary/Secondary
  and Tertiary–Hexagonal Phase pages and epguides. Only the dates are known, so every time is noon UTC,
  just to put the episodes in order. Fit the Seventh was the 1978 Christmas special, and the
  Secondary Phase's other five went out nightly, 21–25 January 1980.
- **Extras, as episodes 13 and 14** (after Fit the Twelfth, before Fit the Thirteenth):
  - "An Interview with Douglas Adams", the Secondary Phase CDs' untitled 50-minute bonus
    ("Chapter 7" in the library). Its date, 26 January 1980, is made up to put it after Fit the Twelfth.
  - "Douglas Adams’s Guide to the Hitchhiker’s Guide to the Galaxy" (55 min, narrated by Peter Jones),
    from the Primary Phase CDs, dated 5 March 1998, when a 40-minute version was first broadcast
    for the 20th anniversary.

  So episode numbers run past Fit numbers from Fit the Thirteenth (episode 15) on.
- **Left out:** the Primary Phase's 16-second "Opening Credits".

## How it was run

```sh
go run ./cmd/podcast-import --dry-run podcasts/hitchhikers/episodes.csv
go run ./cmd/podcast-import podcasts/hitchhikers/episodes.csv
```
