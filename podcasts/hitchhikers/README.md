# The Hitchhiker's Guide to the Galaxy

**Source:** the six radio phases in the audiobook library, CD releases tagged with Picard:
`Douglas Adams/The Hitchhiker’s Guide to the Galaxy <Primary … Quintessential> Phase (Audio Drama)`.
One AAC file per Fit (episode), 44.1 kHz, cover embedded, except the Hexagonal Phase:
one 3 h 40 m Audible file at 22 kHz, with no chapter marks, in a folder misnamed
`The Hitchhiker’s Guide to the Galaxy Hexagonal Phase (Eoin Colfer)`.

**Result:** the 32 Fits and two extras as one 34-episode show,
`/mnt/prudence/podcasts/The Hitchhiker’s Guide to the Galaxy/`, audio copied unchanged, made with `podcast-import` from `episodes.csv`.

## Decisions

- **One show** for all six phases, episodes titled "Fit the First" … "Fit the Thirty-Second".
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
- **Hexagonal Phase** (episodes 29 to 34, added after the first run): its single file is cut into
  six with the manifest's `start` and `end` columns, at the middle of the five silences of over
  2 seconds that ffmpeg's `silencedetect` (-40 dB) found. They gave six episodes of 35 to 39 minutes,
  longer than the 30-minute broadcasts, so presumably this is a longer edit than went out on the radio.
  Dates are the Radio 4 broadcasts, Thursdays at 6:30pm from 8 March 2018 (British Comedy Guide).
  The earlier 28 episodes keep "of 28" in their track numbers, which nothing shows.
- **Left out:** the Primary Phase's 16-second "Opening Credits".

## How it was run

```sh
go run ./cmd/podcast-import --dry-run podcasts/hitchhikers/episodes.csv
go run ./cmd/podcast-import podcasts/hitchhikers/episodes.csv
```
