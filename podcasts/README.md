# Podcasts

Radio series that read better as podcasts than as audiobooks: one show, one episode per broadcast,
in `/mnt/prudence/podcasts/<show>/`, which iTunes reads directly.

Each show has a folder here with a README covering what the source looked like, what came out,
and the decisions made. Most also have the `episodes.csv` manifest that `podcast-import` made
them from (see `cmd/podcast-import`), run on the server from the repo root:

```sh
go run ./cmd/podcast-import --dry-run podcasts/hitchhikers/episodes.csv
go run ./cmd/podcast-import podcasts/hitchhikers/episodes.csv
```

| Show                  | Source                                                     | How                             |
| --------------------- | ---------------------------------------------------------- | ------------------------------- |
| `cabin-pressure/`     | 14-CD box set -> 28 episodes (ALAC), A–Z order             | one-off script, since removed   |
| `bleak-expectations/` | Audible download (30 chapters) -> 30 episodes, re-titled   | one-off script, since removed   |
| `hitchhikers/`        | Five radio phases (one file per Fit) -> 28 episodes        | `podcast-import`                |

The first two were one-off Go programs, written before `podcast-import` existed. Their sources
have gone from the library, so the programs were deleted; they're in git history before the
move from `one-off/` to `podcasts/`.
