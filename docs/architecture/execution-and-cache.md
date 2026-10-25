# How tasks run, and how the cache decides

This page follows a task from `smith run build` to a result on disk: how the
request becomes a graph, how the scheduler runs it, what exactly goes into a
cache key, where results are stored, how a remote cache is trusted, and how
`smith why-miss` and `smith affected` fit in. The code is in `apps/smith`:
`taskgraph`, `taskrun`, `taskcache`, `remotecache`, `affected`, `whymiss`.

## 1. From a request to a task graph

`smith run build test -p web` names **tasks** and **projects**. The selection
(`projectgraph.Select`) turns `-p` patterns into project names: names, globs,
`tag:x`, `kind:x`, `lang:x` and `!pattern` to exclude. `--affected` narrows the
result to the projects a change can break (section 7).

`taskgraph.Build` then adds the tasks those tasks depend on. A task lists its
dependencies in `dependsOn`:

| Entry | Meaning |
|---|---|
| `build` | the `build` task of the **same** project |
| `^build` | the `build` task of every project this one **imports** (its direct dependencies; a dependency without that task is skipped and *its* dependencies are used instead) |
| `api:build` | a named task of a named project |

Tasks are identified as `project:task`. A project that has no task of a
requested name is skipped silently (not every project has a `typecheck`); if
**no** selected project has any, the command fails. The graph is checked for
cycles, and ordered topologically.

A task can also **serve** an app (`serve = "web:start"`): the served task is not
a node in the graph (the runner starts and stops it around the task, see
section 3), but whatever *it* depends on is added as a dependency of the task
that serves it.

## 2. The scheduler

`taskrun.Run` is deliberately small. It starts one goroutine per task and
gives each a `done` channel; a task waits on the channels of its dependencies,
then takes a slot from a counting semaphore (default: the number of CPUs,
`--parallel N` to change it). So independent branches of the graph run
together, a task never starts before its dependencies finished, and the number
of running processes never exceeds the limit.

After its dependencies finish, a task is **skipped** (not failed) if any of them
failed or was skipped, with the reason "because `api:build` did not succeed".
With `--bail`, the first failure also cancels everything that has not started.
The summary at the end counts succeeded, from cache, failed and skipped tasks,
and the time saved by cache hits (the duration of the original run).

### Running a process

A task's `command` is an **argv array, never a shell string**: nothing is
quoted, expanded or glob-matched by a shell, so a command means the same on
Windows and Linux. Before it runs smith expands three placeholders:
`{workspaceRoot}`, `{projectRoot}` and `{smith}` (the path of the running
smith, so a convention's task runs the same version that scheduled it).

The program is looked up on the `PATH` the child will see. smith puts the
workspace's Python virtual environment first on `PATH`, as npm puts
`node_modules/.bin` first for scripts. The environment is the caller's plus
`SMITH_TASK`, `SMITH_TASK_KEY`, `SMITH_WORKSPACE_ROOT`, the task's declared `env`
(placeholders expanded), and `E2E_BASE_URL` for a task that serves an app.

Output is read line by line and prefixed with `project:task │ ` so parallel
output stays readable. On a cache hit the stored log is replayed with the same
prefix.

**Cancelling kills the whole tree.** A build tool starts compilers, which start
workers; killing only the child would leave them running.

- On Unix the child is started in its own process group; on cancel smith sends
  `SIGTERM` to the group, and `SIGKILL` three seconds later.
- On Windows the child is placed in a **Job Object** with
  `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`, so closing the job kills every process
  in it, grandchildren included, even if smith itself dies. The child is also
  started in a new process group so `Ctrl+C` reaches smith first and smith
  decides how to stop the tree.

## 3. Serving an app around a task

When a task sets `serve`, the runner (after a cache miss, never on a hit)

1. starts the served task in the background, its output prefixed with *its*
   id, in its own process tree;
2. polls `serve_url` every 200 ms until it answers with a 2xx. If the app exits
   first, or `serve_timeout` (default 60 s) passes, the task fails with the
   reason and the app is stopped;
3. runs the task with `E2E_BASE_URL` set;
4. stops the app and its tree whatever the result, and waits until it is gone.

The served app is part of the task's cache key (section 4) so a cached
passing suite is a miss when the app changed.

## 4. The cache key

A key is the SHA-256 of a **length-prefixed canonical encoding** of everything
that can change a task's result. Every collection is sorted, so the same
inputs always give the same key, on any machine.

| Part | What it is | Why |
|---|---|---|
| key schema | `smith-task-key-v1` | a change to the encoding bumps it so old entries can never be mistaken for new ones |
| task id | `web:build` | |
| platform | `linux/amd64` | a binary built on one OS is not another's |
| tool | `1.62.0/node@1.62.0+jest@1.62.0` (smith's version, then the project's language plugin and the other plugins that shape it, each with its version) | upgrading smith or any plugin that shaped the project invalidates its results |
| toolchain (if reported) | `node 20.11.0` | a build cached on Node 20 is a miss on Node 22 |
| serve (if set) | `web:start http://127.0.0.1:5173` | the app under test |
| command | the argv **before** placeholders are expanded | so keys do not depend on where the repository is checked out |
| env | the declared `env` | |
| files | path + SHA-256 of **every input file** | |
| upstream | for every project it transitively depends on, a digest of that project's source files | a library change invalidates its dependents |
| deps | the **keys** of the tasks it depends on | an upstream rebuild with different results propagates |

**Inputs.** If a task declares `inputs` (globs relative to the project,
`{workspaceRoot}/…` for root files, `!` to exclude) only those files are
hashed; otherwise every tracked or untracked-not-ignored file of the project.
The file list comes from git, so `.gitignore` is respected. Declared outputs
(under `dist/`) are never inputs. File hashes are computed once per run and
memoised; after a task finishes the memo is dropped for files it may have
changed.

Two properties follow:

- **Correctness over cleverness.** The key includes everything. If you declare
  too few `inputs`, smith can replay a stale result; if too many, it misses more
  than needed. `smith doctor` and `smith why-miss` exist to find both.
- **Fields are added to the key only when set.** A new field (toolchain, serve)
  is written only when non-empty, so adding a feature does not invalidate every
  existing cache entry.

### What a cached result is

On success, `Save` stores:

- every file under the task's declared **output directories** (outputs must be
  directories; a bare file path is rejected so a restore can replace the
  directory wholesale),
- the task's combined log,
- the inputs the key was computed from (for `why-miss`),

and records the key as the task's **last** run.

On a hit, `Restore` **replaces the output directories** with the cached files
and replays the log. On a miss, smith first clears the outputs, so a stale file
from an earlier build can never leak into the new result. This is why two tasks
must not declare the same output directory: restoring one deletes what the other
wrote (`smith doctor` reports it).

Only tasks with `cache = true` and exit code 0 are stored. Failures are never
cached.

## 5. The store

```
.smith/cache/v1/
  ac/<aa>/<key>      action cache: one small JSON entry per key
  cas/<aa>/<sha256>  content-addressed blobs: files and logs, shared between entries
  last/<aa>/<sha256 of task id>   the key of the task's most recent cached run
  tmp/               same volume, so writes land by rename
```

`<aa>` is the first two characters of the name, a shard directory. `[cache] dir`
(or `SMITH_CACHE_DIR`) moves the root.

This is the layout of Bazel's remote cache: entries point at blobs by digest, so
an identical file produced by a hundred tasks is stored once. Writes land by
rename, so an interrupted run never leaves a half-written entry; an entry whose
blob is missing is simply a miss. `Storage` is an interface
(`Has/Get/Put/Delete/Walk/Touch`) with a conformance test that every backend
passes.

**Pruning.** `taskcache.Prune` keeps the store under a size: the least recently
used *entries* go first until the blobs the remaining entries reference fit, then
every blob no entry references is swept. Using an entry (`Touch`) counts as use.
The size comes from `[cache] max_size_mb` (default 2048; 0 means no limit). After a
run that stored new results, `taskrun` calls it for the **local** store through
`cachemaintenance.AfterRun`, which walks the store at most about once an hour (a small
state file keeps the running size) and leaves anything used in the last five minutes. A
failure is a warning, never a failed run. `smith cache prune` runs it on demand. How it is
wired, with a sequence diagram, is in
[The cache key, hits and misses](cache.md#pruning).

## 6. The remote cache

CI fills a remote store and developers read it. `remotecache.Tiered` puts the
remote **behind** the local store:

- *read:* local first; on a local miss the remote is asked, the object is
  **verified**, and a good one is copied into the local store;
- *write:* both, unless the session is read-only. A session is read-only when
  `[cache] remote_read_only` or `SMITH_CACHE_READ_ONLY` says so, or when the CI
  environment says the run is for a pull request (`GITHUB_EVENT_NAME` is
  `pull_request`, `BUILD_REASON` is `PullRequest`, or `CI_PIPELINE_SOURCE` is
  `merge_request_event`); any other run writes. Without `SMITH_CACHE_KEY` the
  remote is switched off with a warning, because nothing is shared unsigned.

Trust is the central design point. The remote is a place anyone with write
access to the bucket could poison, so **every object is signed** with an
HMAC-SHA256 over its kind, key and content, using a secret only the team holds
(`SMITH_CACHE_KEY`). The name is part of the signed message, so an object cannot
be replayed under another key; a CAS blob is also checked against its own
digest. An object that fails verification, an unreachable store, a timeout: all
are a **miss**, reported once as a warning. A remote cache can make a run
slower; it can never make it wrong.

Backends (`remotecache.Open`, chosen by the URL scheme of `[cache] remote` or
`SMITH_CACHE_REMOTE`): an HTTP object store (`https://`, optional bearer token in
`SMITH_CACHE_TOKEN`), S3 (`s3://`, AWS Signature V4, implemented in
`sign_s3_request_algorithm.go`, so no SDK; needs `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`), Azure Blob (`azblob://`, SAS token in
`AZURE_STORAGE_SAS_TOKEN`).
`last/` stays local: it exists to explain misses on *this* machine.

Credentials reach the generated pipelines as secrets: `smith ci check` lists
what each remote scheme needs, and only the steps that run tasks receive them.

## 7. Affected

`smith affected` answers "which projects can this change break?" cheaply and
conservatively.

1. Pick a base: `--base`, else the pull-request target the CI provider exports
   (`SMITH_BASE`, `GITHUB_BASE_REF`, `SYSTEM_PULLREQUEST_TARGETBRANCH` or
   `CI_MERGE_REQUEST_TARGET_BRANCH_NAME`, in that order), else `[release] branch`.
2. `git diff` against the **merge base** with it (`origin/<base>` is tried first),
   plus uncommitted and untracked files.
3. Each changed file belongs to the project whose root contains it (the deepest
   one when roots nest); a shared schema file also belongs to the project that
   owns it. A file that belongs to no project and is not neutral (`*.md`,
   `docs/**`, `LICENSE`, `.github/ISSUE_TEMPLATE/**`, `.vscode/**`,
   `*.code-workspace`) or excluded by `[workspace] exclude` is a
   **workspace-level** change (`smith.toml`, a lockfile, a pipeline) and affects
   everything. [How affected projects are computed](affected.md) has the detail.
4. Add every project that depends on an owner, transitively.

Whenever smith cannot be sure — no base, an unknown base, a shallow clone, no
common ancestor — it says so and affects **everything**. Too wide costs
minutes; too narrow reports green for code nothing checked.

The same graph gives `smith affected --paths`: for each app, its directory and
those of everything it depends on, which is what a deployment platform's "watch
paths" need.

## 8. Explaining a miss

Each entry stores the key's inputs. `smith why-miss project:task` recomputes the
key as a run would now (`taskrun.KeyOf`), loads the inputs of the task's last
cached run, and prints the **difference**, in a stable order: tool, toolchain,
serve, platform, command, env, files added / removed / changed (by path),
upstream projects whose sources changed, dependency tasks whose keys changed.
It follows the chain: a changed dependency is itself explained with
`smith why-miss core:build`. It runs nothing.

## 9. Things deliberately not done

- *Sandboxing / hermetic builds.* smith trusts a task's declared inputs and does
  not run it in a sandbox. That is what lets it run `npm`, `dotnet` and `cargo`
  unmodified; the price is that wrong `inputs` give wrong cache hits.
- *Distributed execution.* Tasks run on one machine; only results are shared.
- *File watching.* smith recomputes from git on each run (`smith dev` starts
  your dev servers; it does not rebuild).
