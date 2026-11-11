<!-- lore-master
parent: cache.md
-->
# smith cache push

`smith cache push` uploads every entry of the local task cache that the remote cache does not
hold, with the files the entry stores (its log and its outputs). It signs each object with
`SMITH_CACHE_KEY` exactly as a run does. It asks the remote first, with one `HEAD` request for
each entry and each file, and never uploads what the remote holds. With `--dry-run` it asks and
uploads nothing.

A run uploads what it stores at once. An upload that fails is a warning, and the next run does
not send it again, because it finds the entry in the local cache. Push sends those entries.

## When to use it

- A run warned `the remote cache could not be written` (the remote did not answer, or this
  machine was offline).
- Home in the interactive shell says `Remote cache: connected · 12 local entries not uploaded`.
- You set up a new remote and want it filled from this machine now.

## Synopsis

```sh
smith cache push [--dry-run] [--json]
```

## Arguments and flags

<!-- smith:begin cache push -->
Upload every local cache entry the remote cache does not hold

Usage: `smith cache push [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | show what would be uploaded; upload nothing |
| `--json` | boolean | off | print the result as JSON |

<!-- smith:end cache push -->

- The order is safe for a reader on another machine: the files of an entry go before the
  entry, so the remote never holds an entry whose files are missing.
- Four entries go at once. A failed entry does not stop the others. Run the command again: it
  uploads only what is still missing.
- A remote that refuses the credentials (`401` or `403`) stops the push at the first refusal.
- A read-only remote (`[cache] remote_read_only`, `SMITH_CACHE_READ_ONLY`, or a pull-request
  build) is refused, except with `--dry-run`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every entry the remote lacked was uploaded, there was nothing to upload, or the dry run printed. |
| 1 | No remote, no `SMITH_CACHE_KEY`, a read-only remote, the remote refused the credentials or did not answer, or an entry failed. |

## Files and environment

Reads the cache directory (`[cache] dir` or `SMITH_CACHE_DIR`). Variables: `SMITH_CACHE_REMOTE`,
`SMITH_CACHE_KEY`, `SMITH_CACHE_READ_ONLY` and the credentials of the store (`SMITH_CACHE_TOKEN`,
the `AWS_*` variables, or `AZURE_STORAGE_SAS_TOKEN`; see
[Remote cache](../reference/tasks-and-cache.md#remote-cache)). It starts no plugins.

## Examples

These samples ran against a local HTTP object store (a test server that takes PUT, GET and
HEAD), on this repository.

```sh
smith cache push --dry-run
```

```text
Asking http://127.0.0.1:18731/team which local entries it lacks (dry run: nothing is uploaded).
  ○ plugin-go:slice-check  would upload (0 files, 13.8 KB)
  ○ plugin-sdk:slice-check  would upload (1 file, 29.6 KB)
  ○ smith:slice-check  would upload (1 file, 343.6 KB)
✓ Dry run: 3 of 5 entries would be uploaded to http://127.0.0.1:18731/team (2 files, 387.0 KB); 2 already on the remote. Nothing was uploaded.
→ run smith cache push to upload them
```

`0 files` means the remote holds the entry's files already (another entry stores the same log).

```sh
smith cache push
```

```text
Uploading the local entries that http://127.0.0.1:18731/team lacks.
  ✓ plugin-sdk:slice-check  uploaded (1 file, 29.6 KB)
  ✓ plugin-go:slice-check  uploaded (0 files, 13.8 KB)
  ✓ smith:slice-check  uploaded (1 file, 343.6 KB)
✓ Uploaded 3 entries to http://127.0.0.1:18731/team (2 files, 387.0 KB); 2 already on the remote.
```

Again, with nothing left:

```text
Uploading the local entries that http://127.0.0.1:18731/team lacks.
✓ http://127.0.0.1:18731/team holds all 5 local entries: nothing to upload.
```

`--json` prints the counts and one step for each entry (`uploaded`, `would-upload`,
`present` or `failed`):

```json
{
  "store": "http://127.0.0.1:18731/team",
  "dryRun": false,
  "entries": 5,
  "uploaded": 0,
  "present": 5,
  "failed": 0,
  "blobs": 0,
  "bytes": 0,
  "steps": [
    { "key": "7f04f10a…", "task": "actions:slice-check", "outcome": "present" }
  ]
}
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `× no remote cache is set for this workspace` | Run `smith cache setup`. |
| `× the remote cache <store> is off: SMITH_CACHE_KEY is not set` | Set the team's key. smith uploads nothing unsigned. |
| `× the remote cache <store> refused the credentials (it answered 403 Forbidden)` | The token or the key of the store is wrong, expired, or lacks a right. The remedy line names the variables of the store. |
| `× the remote cache <store> is read-only here` | Unset `SMITH_CACHE_READ_ONLY` or `remote_read_only`. A pull-request build never writes: push from the main branch. |
| `× <task>  not uploaded: <reason>` | That entry failed; the others went. Run the command again. |

## Related

[`smith cache`](cache.md), [`smith cache status`](cache-status.md),
[`smith cache setup`](cache-setup.md), [Remote cache](../reference/tasks-and-cache.md#remote-cache).
