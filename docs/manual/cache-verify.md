<!-- lore-master
parent: cache.md
-->
# smith cache verify

`smith cache verify` checks the local task cache for damage. It reads every blob and checks
that it hashes to its name; reads every entry and checks that it can be read and names only
blobs that exist; looks for blobs no entry uses; and looks for half-written files left in
`tmp/`. With `--fix` it removes what is broken. It exits 1 while problems remain.

smith checks a blob against its digest when it stores one and when it fetches one from the
remote cache, but not each time it restores a local one. A blob that was damaged on disk would
therefore be restored as it is, which is why this check exists. An entry whose blob is gone is
only a cache miss, but it never hits again until it is removed.

## When to use it

- After a crash, a power cut or a full disk while a build was running.
- When [`smith doctor`](doctor.md) warns that the local cache has unusable entries.
- When a task you expect to hit always runs, and `smith why-miss` shows no change.

It reads the whole cache, so it takes a moment on a large one. [`smith cache status`](cache-status.md)
and `smith doctor` do a cheaper look that hashes nothing.

## Synopsis

```sh
smith cache verify [--fix] [--json]
```

## Arguments and flags

<!-- smith:begin cache verify -->
Check the local cache for damage; --fix removes what is broken

Usage: `smith cache verify [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--fix` | boolean | off | delete what is broken: corrupt blobs and the entries that name them, unused blobs, leftover files |
| `--json` | boolean | off | print the report as JSON |

<!-- smith:end cache verify -->

What it reports, one line per problem:

| Kind | Meaning | What `--fix` does |
|---|---|---|
| `corrupt-blob` | A blob's content does not hash to its name. | Deletes the blob. |
| `damaged-entry` | An entry names a corrupt blob. | Deletes the entry. |
| `missing-blob` | An entry names a blob that is gone. | Deletes the entry. |
| `unreadable-entry` | An entry file is not a cache entry. | Deletes it. |
| `orphan-blob` | No usable entry names the blob. | Deletes the blob. |
| `leftover-tmp` | A file in `tmp/` from a write that did not finish. | Deletes it. |

A blob written, and a file in `tmp/`, within the last minute may belong to a write in progress
in another smith process, so they are not reported and `--fix` leaves them. Without `--fix`
nothing is changed. With `--fix`, each line ends in `(removed)` when the file was deleted.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No problems, or `--fix` removed all of them. |
| 1 | Problems remain (always, without `--fix`), not in a workspace, or the store could not be read. |

The report is printed in either case; the `✖` line with the remedy goes to standard error.

## Files and environment

Reads every file in the cache directory; with `--fix` deletes the broken ones. Variable:
`SMITH_CACHE_DIR`. The remote cache is not checked: every object fetched from it is verified
against its signature and digest when it is fetched.

## Examples

A healthy cache:

```sh
smith cache verify
```

```text
Checked 5 entries and 4 blobs in ~/work/acme/.smith/cache: no problems.
```

After a blob was overwritten (the bytes of `core:bundle4`'s output replaced by `bit rot`):

```sh
smith cache verify
```

```text
Checked 5 entries and 4 blobs in ~/work/acme/.smith/cache: 2 problems.
  corrupt-blob     0e9309b05cf8…  its content hashes to 28d39a931605…, not to its name
  damaged-entry    ba51ebc972a7…  the entry for core:bundle4 names blob 0e9309b05cf8…, which is corrupt
✖ the cache at ~/work/acme/.smith/cache has 2 problems
  → run `smith cache verify --fix` to remove the damaged files; the next run computes what they held again
```

```sh
smith cache verify --fix
```

```text
Checked 5 entries and 4 blobs in ~/work/acme/.smith/cache: 2 problems.
  corrupt-blob     0e9309b05cf8…  its content hashes to 28d39a931605…, not to its name (removed)
  damaged-entry    ba51ebc972a7…  the entry for core:bundle4 names blob 0e9309b05cf8…, which is corrupt (removed)
```

```sh
smith cache verify
```

```text
Checked 4 entries and 3 blobs in ~/work/acme/.smith/cache: no problems.
```

`--json` prints `dir`, `entries`, `blobs` and `problems`, each with `kind`, `key`, `detail` and
`fixed`.

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ the cache at <dir> has 2 problems` | Run `smith cache verify --fix`. The next run of the affected tasks runs them again. |
| `✖ could not read the cache at <dir>` | A permissions or disk error. Fix it, or start over with `smith cache clear`. |

## Related

[`smith cache`](cache.md), [`smith cache status`](cache-status.md),
[`smith cache clear`](cache-clear.md), [`smith doctor`](doctor.md),
[The cache key, hits and misses](../architecture/cache.md).
