# ADR 0096: The remote cache can be cleared, from the command line and the shell

- Status: accepted, implemented
- Date: 2026-11-11

## Context

`smith cache clear` deleted the local cache only. A remote that held results nobody wanted (a
bad toolchain cached, a remote being left) could be emptied only with the cloud's own tools,
and smith's remote store could not list or delete. The SAS token that `smith cache setup` makes
had no delete right (`rwcl`).

## Decision

The owner made these decisions on 2026-11-11.

1. **Remote clear deletes the objects.** `smith cache clear --remote` deletes this workspace's
   objects on the remote: every `ac/` and `cas/` object under the remote's prefix.
   `smith cache clear --all` clears the local cache and the remote. Plain `smith cache clear`
   stays local and makes no network request.
2. **The question names what goes.** It names the store (host and prefix, never credentials or
   a query), the workspace and the count, and says that the team and CI share the cache.
   `--yes` skips it; without a terminal and without `--yes` the command refuses with the
   remedy. `--dry-run` prints the same question and deletes nothing.
3. **List and delete on every store.** `remotecache.ObjectStore` gains `List` and `Delete`. S3
   uses `ListObjectsV2` with the prefix and a continuation token, and `DeleteObject` (one
   `DELETE` for each object, 8 at a time; `DeleteObjects` needs a `Content-MD5` body and gains
   little at cache sizes). Azure Blob uses List Blobs with a marker and Delete Blob, with
   `x-ms-version: 2021-08-06`. A plain `https://` store cannot list: smith deletes there only
   the keys the local cache holds, says `up to N objects`, and says that the rest stays.
4. **Rights.** A 401 or 403 on a list or a delete says that the credentials need list and
   delete rights, and how to get them for the store. `smith cache setup` makes the Azure SAS
   token with `rwdcl` (delete added); a token made before has no delete right and must be
   issued again.
5. **The shell.** The `Clear cache` button asks which cache goes (local, remote or both) when
   the workspace has a remote, then shows the confirmation with the text of
   `smith cache clear … --dry-run`, so the shell and the command line say the same thing.
   Without a remote it clears the local cache as before. VS Code and JetBrains do not change.

## Options considered

- **Delete by prefix in one call.** Neither S3 nor Azure Blob deletes a prefix. A lifecycle
  rule would, but hours later and outside smith.
- **A remote clear that only forgets** (a new key prefix, a generation number). The objects
  would stay and cost storage, and the owner chose to delete them.
- **List on the plain store through a convention** (an index file). Every server would need it.
  Deleting the keys smith knows, and saying so, works with any server.

## Consequences

- An Azure SAS token from an older setup refuses the clear with a 403 and the remedy.
- A plain `https://` store keeps the objects that only other machines stored.
- A clear in progress that stops (a network failure) leaves the rest; running it again deletes
  what is left.
