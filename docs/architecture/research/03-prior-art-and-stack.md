# Research 03 — Prior art and recommended stack

Marks: [V] verified against the cited source, [M] from memory (verify before relying on it).

## Please

- Target hash combines pre-build rule hash, post-build rule hash, config hash, source hash, secret hash [V:
  github.com/thought-machine/please `src/build/incrementality.go`]. Rule hash covers label, sorted deps, outputs,
  command, flags, pass-env.
- Caches: directory, HTTP (plain PUT/GET), command-driven (`CACHE_KEY`) [V: please.build/cache.html]. Artifacts
  stored only after success.
- `plz query changes --since <rev>` [V: please.build/commands.html]. First-class plugins: Go, Python, Java, C/C++,
  Shell, Proto [V]. No JS/TS, .NET or Dart plugins [V]. No native Windows [M].

## Cache layout and `Storage`

- Bazel HTTP cache: `GET/PUT /ac/{sha256}` and `/cas/{sha256}` [V: bazel.build/remote/caching].
- moon: `.moon/cache/{blobs,manifests,outputs,locks,states}`, hydration on hit [V].
- Turborepo: HTTP API, optional HMAC-SHA256 signing, failed verification = miss [V].

Recommended layout (design):

```
<cache>/v1/ ac/<aa>/<sha256-key>   JSON manifest {schema, outputs[{path,mode,digest,size}], stdout/stderr digests, exit}
            cas/<aa>/<sha256>      blob (digest of uncompressed bytes)
            tmp/                   same volume, for atomic rename
            VERSION
```

Key = SHA-256 over a canonical, versioned, length-prefixed encoding of: schema version, task id, command/args,
sorted input digests, declared env, tool versions, dependency task keys, platform (opt-in). Never absolute paths
or timestamps. Atomic: write tmp → fsync → rename (retry with backoff on Windows). Verify digest on read; corrupt
= miss. Prune by size/age with mark-and-sweep from AC roots; use an explicit last-used touch (atime unreliable).

`Storage`: `Has/Get/Put(kind, digest, …)`, `Delete`, optional `Walk`, `Capabilities()`. Compression and signing
are decorators. Mirroring Bazel AC/CAS lets bazel-remote or an HTTP server be a future backend.

## Plugin protocol

go-plugin: handshake line + gRPC, Go-centric, unix-socket friction on Windows [V]. Terraform: gRPC with major
protocol versions, additive minors [V]. **Recommendation: JSON-RPC 2.0 over stdio (LSP framing)**:
`initialize` (host: protocolVersions, workspaceRoot; plugin: protocolVersion, kind, capabilities),
`initialized`, `shutdown`, `exit`, `$/cancelRequest`, progress notifications; stderr is log-only; integer major
version, additive via capability flags; methods not advertised are never called. `plugin.json` manifest for
discovery without spawning. Batch calls; avoid per-file RPC.

## Go libraries

TUI: Bubble Tea + Lip Gloss + Bubbles [M: check current major/import paths]; always provide a non-TTY mode.
CLI: cobra (or kong). Conventional commits: small in-house parser with a spec corpus (or go-conventionalcommits).
Git: shell out to `git` behind an interface (`-z`, `--no-optional-locks`). Globbing: doublestar v4. Hash: SHA-256.
Windows processes: Job Objects with KILL_ON_JOB_CLOSE, `CREATE_NEW_PROCESS_GROUP` for Ctrl+C, explicit `.cmd`
resolution, long-path handling.

## Distribution

- esbuild: per-platform npm packages as `optionalDependencies` with `os`/`cpu`, JS shim resolves and execs,
  fallback postinstall download with SHA-256 [V].
- ruff/uv: maturin platform wheels with the binary in `scripts/` [M]. Biome: same npm pattern [M].
- GoReleaser OSS for archives/checksums/GitHub release; its npm publishing is Pro-only and download-based [V].
  Emit npm packages and wheels from GoReleaser's dist with a small script (or `gowheels`, unverified).

## Top risks

1. Cache key correctness (undeclared inputs). 2. Windows filesystem (rename failures, long paths, case,
symlinks, CRLF). 3. Windows process trees/signals. 4. Plugin protocol lock-in. 5. Affected in shallow/sparse
clones. 6. Cache security/poisoning. 7. Distribution complexity (skipped optionalDependencies, musl/glibc,
signing). 8. Scope and Nx-parity creep.
