<!-- lore-master
parent: cache.md
-->
# smith cache key

`smith cache key` makes the key a remote cache needs, `SMITH_CACHE_KEY`, and puts it on the
clipboard without printing it. The key is 32 bytes from the operating system's secure random
source, shown as 64 hex characters. It lives in memory while the command runs. smith writes it
to no file, no log, not `smith.toml`, and not the operation journal.

## What the key is for

smith signs every object it stores in the remote cache with this key (HMAC-SHA256) and checks the
signature before it uses an object. Without the key the remote cache is off. You make the key
once, give the same value to every pipeline and every developer who uses the remote, and keep it
in the environment only. The whole guide, with rotation and the SAS token, is in
[Set up the remote cache](../reference/tasks-and-cache.md#set-up-the-remote-cache).

## Synopsis

```sh
smith cache key [--print] [--clear-after <duration>]
```

## Arguments and flags

<!-- smith:begin cache key -->
Make a key (SMITH_CACHE_KEY) for the remote cache and copy it, unseen

Usage: `smith cache key [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--clear-after <clear-after>` | duration (`60s`, `2m`, `0`) | `60s` | clear the clipboard after this long if it still holds the key; 0 leaves it |
| `--print` | boolean | off | print the key instead of copying it |

<!-- smith:end cache key -->

## On a terminal

The command copies the key, then waits for one key press, with no Enter:

| Key | Effect |
|---|---|
| `c` | Copies the same key again and restarts the clearing timer. |
| `g` | Makes a new key and copies it. It replaces the key shown before: use the newest one everywhere. |
| `u` | Sets the key as a user environment variable. smith shows the exact command and asks `y` first. On Windows it runs PowerShell with the key on standard input. On other systems it shows the line for your shell profile and edits nothing. |
| `q` | Ends. The clipboard is cleared now if it still holds the key. |

The clipboard is written with `clip`, `pbcopy`, `wl-copy`, `xclip` or `xsel`, whichever is
installed, else with the terminal's own copy request (OSC 52), which works over SSH when the
terminal allows it. OSC 52 cannot be read back, so that copy is never cleared for you. When no
clipboard exists, smith says so and offers `p` to print the key here.

The clipboard is cleared after `--clear-after` (60 seconds), but only when it still holds the
key, so something you copied since stays. `--clear-after 0` turns this off. A clipboard-history
tool may keep a copy of the key. Without a terminal, the command copies once and ends, and the
clipboard is not cleared for you.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The key was copied or printed. |
| 1 | The random source failed, or there is no clipboard and no terminal (use `--print`). |

## Examples

```sh
smith cache key
# Key copied to the clipboard (not shown). Paste it where you need it.
```

```sh
smith cache key --print | gh secret set SMITH_CACHE_KEY
```
