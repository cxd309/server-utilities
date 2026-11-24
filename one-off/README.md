# One-off jobs

Scripts for unusual releases that the regular tools don't handle, each run once by hand on the server.
They're kept as worked examples for the next odd one to remember what I did.

Each job has its own folder with a README covering what the source looked like, what came out,
the decisions made, and the command that was run.

| Job               | What it did                                            |
| ----------------- | ------------------------------------------------------ |
| `cabin-pressure/` | 14-CD box set -> 28 podcast episodes (ALAC), A–Z order |
| `bleak-expectations/` | Audible download (30 chapters) -> 30 podcast episodes, re-titled |

Run from the repo root, e.g. `go run ./one-off/cabin-pressure --dry-run`.
If two jobs ever need the same code, that's the sign to turn it into a proper tool.
