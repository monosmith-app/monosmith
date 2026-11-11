# ADR 0095: The remote cache says what it did, and a person can see and fill it

- Status: accepted, implemented
- Date: 2026-11-11

## Context

The owner cleared the local cache with `smith cache clear` and still got 31 hits. They came from
the remote cache, but smith printed the same `✓ from cache` for both tiers. Nothing showed
whether the remote answered, how much of the local cache it held, or a way to send it what a
failed upload left behind: `Tiered.Put` uploads in the same call, and an upload that fails is a
warning that no later run repeats.

## Decision

The owner made these decisions on 2026-11-11.

1. **Hit label.** A hit the remote served prints `<task> ✓ from remote cache`. The closing line
   counts them: `(31 from cache, 29 of them remote, …)`. `--json` adds `source` (`local` or
   `remote`) to each cached result. `--json-events` adds `source` to `task.finished` and
   `cachedRemote` to the `run.finished` summary. No existing field changes. The TUI run view
   says `from remote cache` in the task's row and in the closing line. The runner asks
   `taskcache.SourceOf` before the lookup: an entry the local tier lacks comes from the remote.
2. **Connection state in the TUI header.** The header ends with the cache mark and a state mark
   from the theme icons: `≈✓` connected, `≈▲` no answer, `≈×` the credentials are refused or
   missing. ASCII terminals get `[c][v]`, `[c][!]`, `[c][x]`. The probe is one `HEAD` of an
   object that does not exist, run in the background with a 5 second timeout. The header says
   `≈ checking…` until it ends. Without a remote there is no mark.
3. **Pending count on Home.** Home says how many local entries the remote lacks, with a `Push`
   button. smith asks `HEAD` for each local entry, at most 500 entries, 8 at a time, for at most
   10 seconds. When it stops early it says `at least N`.
4. **`smith cache push`.** It uploads every local entry the remote lacks. The blobs of an entry go
   before the entry. It asks first and never uploads what the remote holds. It has `--dry-run`
   and `--json`, prints one line for each entry it sends and a closing line, and refuses a
   read-only remote. Every remote write goes through `remotecache.Upload`, so a pushed entry is
   signed exactly as a run signs it.

The probe, the count and the push live in a new package, `remotesync`. It opens no session, so
the TUI calls it in the background and the tests run it against an `httptest` server.

## Options considered

- **Mark the hit in `Tiered`** (record which keys it pulled). This needs a lock and state in the
  storage, and a wrapper such as the write meter hides it. Asking the local tier before the
  lookup needs neither.
- **Retry failed uploads at the next run** (a queue of pending uploads). This puts more work on
  every run and still shows nothing. A command that a person runs, and a count that Home shows,
  make the state visible.
- **Count by listing the remote.** A plain `https://` store cannot list, and S3 and Azure lists
  are paged. `HEAD` for each local entry works on every store and is capped.

## Consequences

- A probe on S3 needs list rights on the bucket: without them, S3 answers 403, not 404, for a
  missing object, and smith shows "refuses the credentials".
- A large local cache shows `at least 500` on Home. `smith cache push --dry-run` gives the full
  count.
- The remote clear (`smith cache clear --remote`) follows in its own change and decision.
