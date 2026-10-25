<!-- lore-master
parent: cache.md
-->
# smith cache status

`smith cache status` shows the local task cache: where it is, how many entries and blobs it
holds, its size against `[cache] max_size_mb`, when smith last pruned it, whether a remote
cache is configured and in use, and the five tasks with the largest entries. It reads the
store and changes nothing; it does not even create the folder when nothing has been cached.

## When to use it

- The disk is filling and you want to know whether the cache is why.
- You changed `max_size_mb` and want to see how full the cache is against it.
- A teammate's machine does not share results and you want to see whether the remote cache is
  configured, and whether it is active or switched off.

It does not hash any file; [`smith cache verify`](cache-verify.md) does.

## Synopsis

```sh
smith cache status [--json]
```

## Arguments and flags

<!-- smith:begin cache status -->
Show the local cache: where it is, how big, how full, and the largest entries

Usage: `smith cache status [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the status as JSON |

<!-- smith:end cache status -->

What the rows mean:

- **entries** and **blobs**: results stored, and the content-addressed files they point at.
  Blobs are shared, so there are usually fewer blobs than entries plus outputs.
- **size**: the bytes of entries, blobs and last-run pointers, against the limit and as a
  percentage. With no limit (`max_size_mb = 0`) it says so.
- **last prune**: when smith last pruned, by itself or by `smith cache prune`; `never` before that.
- **remote**: `none`, or the scheme and host of the remote cache and whether it is `active`,
  `active, read-only`, or `off: set SMITH_CACHE_KEY to use it`. Credentials, paths and query
  strings in the address are never printed.
- **largest entries**: per task, the largest cached result and how many entries the task has.
- Two lines with `✖` and an arrow appear when the cache is over its limit, or when some entries
  are unreadable or name a blob that is gone; the arrow names the command that fixes it.

`--json` prints one document: `dir`, `exists`, `entries`, `blobs`, `bytes`, `limitBytes` (0 for
no limit), `usedPercent` (absent with no limit), `lastPrune` (absent before the first),
`brokenEntries`, `largest` (`task`, `entries`, `largestBytes`) and `remote` (`configured`,
`scheme`, `host`, `active`, `readOnly`).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The status was printed, whatever it says. |
| 1 | Not in a workspace, or `smith.toml` does not load. |

## Files and environment

Reads `smith.toml`, the entries and the sizes of the files in the cache directory. Variables:
`SMITH_CACHE_DIR` (another directory), `SMITH_CACHE_REMOTE`, `SMITH_CACHE_KEY` and
`SMITH_CACHE_READ_ONLY` (they decide what the remote row says; their values are never printed).

## Examples

```sh
smith cache status
```

```text
local cache  ~/work/acme/.smith/cache
  entries     5
  blobs       4
  size        6.0 MB of 10.0 MB (60.1% full)
  last prune  2026-10-25 16:51
  remote      none
largest entries
  core:bundle3                 3.0 MB  1 entry
  core:bundle2                 2.0 MB  1 entry
  core:bundle1                 1.0 MB  1 entry
  core:build                      0 B  1 entry
  util:build                      0 B  1 entry
```

With a remote configured but no key, as on a laptop that was never given the team's secret:

```sh
SMITH_CACHE_REMOTE=https://user:pass@cache.example.com/team smith cache status
```

```text
  remote      https://cache.example.com (off: set SMITH_CACHE_KEY to use it)
```

As JSON:

```sh
smith cache status --json
```

```json
{
  "dir": "~/work/acme/.smith/cache",
  "exists": true,
  "entries": 5,
  "blobs": 4,
  "bytes": 6298112,
  "limitBytes": 10485760,
  "lastPrune": "2026-10-25T16:51:03.0168745Z",
  "brokenEntries": 0,
  "largest": [
    { "task": "core:bundle3", "entries": 1, "largestBytes": 3145728 },
    { "task": "core:bundle2", "entries": 1, "largestBytes": 2097152 }
  ],
  "usedPercent": 60.1,
  "remote": { "configured": false, "active": false, "readOnly": false }
}
```

(`largest` is cut to two of its five rows, and the JSON is printed one field per line by smith.)

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ there is no smith workspace here` | Run it inside a workspace, or set `SMITH_WORKSPACE_ROOT`. |
| `✖ the cache is over its limit` with `→ smith cache prune ...` | Not an error: automatic pruning has not caught up yet. Run the command it names. |
| `✖ 2 entries cannot be used` with `→ smith cache verify --fix ...` | Entries that can never be a hit. Run the command it names. |

## Related

[`smith cache`](cache.md), [`smith cache prune`](cache-prune.md),
[`smith cache verify`](cache-verify.md), [`smith info`](info.md), [`smith doctor`](doctor.md).
