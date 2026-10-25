# The docker plugin

The `docker` plugin ships apps as container images. For every app that has a `Dockerfile` it adds
an `image` task, builds and pushes the image on release, checks that the Docker daemon answers, and
writes a `Dockerfile` for a new Go app, static web app or Node app. It is a deploy plugin that
works as a workspace convention for every language.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `deploy` |
| Applies to | `node`, `go`, `python`, `rust`, `dotnet`, `ruby`, `php`, `cpp`, `dart`, `jvm` |
| Capabilities | `generate`, `publish`, `ci`, `doctor`, `contribute` |
| Stacks it publishes | none; turn it on with `[conventions.docker]` or name `docker` in a stack's `deploy` |
| Project kinds | `app` with a `Dockerfile` in its root; libraries and packages do not ship as images |

## What it does

### The `image` task

```text
svc:image — Build the app's container image (ghcr.io/acme/svc:local)
  command    docker build -f apps/svc/Dockerfile -t ghcr.io/acme/svc:local -t ghcr.io/acme/svc:latest .
  cwd        .
  dependsOn  build (this project's build, first)
  inputs     **/* (every file of the project)
  outputs    —
  env        —
  cache      false

Every value above comes from convention docker, except where marked.
```

| Property | Value |
|---|---|
| Command | `docker build -f <app>/Dockerfile -t <image>:local -t <image>:latest <context>` |
| Context | the workspace (`.`), or `[projects.<name>.docker] context` |
| dependsOn | the app's `build`, and its `prune` when it has one (a Node app that builds) |
| Cached | never: the image lives in the Docker daemon, not in files smith can restore |
| Name | the first registry's reference when a registry is set, else the project's lowercase name |
| `verify` | `smith ci verify` builds it for affected apps on pull requests, where Docker is installed, so a broken `Dockerfile` fails the pull request and not the release |

It is not named `build`, so `lint`, `test` and `build` never need a Docker daemon.

### Writing a `Dockerfile`

When a new app is created, the plugin writes a `Dockerfile` (and a `.dockerignore` if the workspace
has none) for:

| App | Dockerfile |
|---|---|
| Go | builds the binary in `golang:<go version>-alpine`, ships on `gcr.io/distroless/static-debian12:nonroot` |
| static web (an `index.html`, `angular.json` or `astro.config.mjs`) | nginx serving `dist/<app dir>` |
| Node that has a `build` script | installs from the app's pruned `dist/<app dir>` with `npm ci --omit=dev` (Node 24) |
| anything else | no file; a note says to write `<dir>/Dockerfile` |

```sh
smith add app svc --lang go
```

```text
Wrote ...
  .dockerignore
  apps/svc/Dockerfile
  ...
smith run image -p svc builds its container image; a release pushes it to the [conventions.docker] registry (or registries) in smith.toml.
```

```dockerfile
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/app ./apps/svc

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
```

## Settings

`[conventions.docker]` for every app, `[projects.<name>.docker]` for one; the project's value wins.
These are the only plugin settings that have their own tables in `smith.toml`.

| Key | Applies to | Means | Default |
|---|---|---|---|
| `registry` | both | where the release pushes images: `ghcr.io/acme` | unset: built, never pushed |
| `registries` | both | the list form: strings, or tables with `url`, `username_env`, `password_env`; the same image and tags go to each. Never together with `registry` | none |
| `push` | both | `false` builds but never pushes | `true` when a registry is set |
| `verify` | both | `false` stops `smith ci verify` building the image | `true` |
| `tag_commit` | both | also tag `sha-<short commit>` | `false` |
| `tag_branch` | both | also tag with the release branch's name | `false` |
| `description_from` | both | a file whose text becomes the Docker Hub repository description | none |
| `short_description` | both | the Docker Hub one-line summary, at most 100 characters | the project's own description |
| `image` | project | the image's name under the registry | the project's name, lowercase |
| `context` | project | the build context | `.` |

`CONTAINER_REGISTRY` (an environment variable) is the deprecated way to name a registry; it is read
when no registry is configured, with a warning.

## Releasing

On release the plugin builds the image once, tags it with the version and `latest` (and the commit
and branch tags if asked), logs in to each registry and pushes, `latest` first and the version last,
so the version tag shows the registry is complete. A re-run pushes only to the registries that
lack the version. With no registry, or `push = false`, the image is built and not pushed and the
release goes on. The release result lists the digest each registry holds.

```text
Dry run: nothing is published or tagged.
  svc                  unreleased → 0.1.0  (1 releasing commit(s))
svc: docker: would push ghcr.io/acme/svc:0.1.0, ghcr.io/acme/svc:sha-bddd387, ghcr.io/acme/svc:latest
```

| Registry | Signs in with |
|---|---|
| `ghcr.io` | `GITHUB_TOKEN` (as `GITHUB_ACTOR`) |
| Docker Hub (`docker.io`) | `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` |
| another | `CONTAINER_USERNAME`, `CONTAINER_PASSWORD`, or the variables its entry names in `username_env` and `password_env`; otherwise the daemon's own login |

An app with a `Dockerfile` releases without a `[release]` section. See
[Releasing](../../../reference/releasing.md#container-images).

## CI

`ci` returns the credentials of every registry the apps push to, and the GitHub permission
`packages: write` for `ghcr.io` (the generated workflow has `packages: write` and passes
`GITHUB_TOKEN`). For a Docker Hub description it adds a post-release step, `smith docker
description`, which runs after every push to the release branch because a docs-only commit releases
nothing. Docker is on the hosted Linux images of GitHub and Azure.

## Doctor

| Check | Fails when |
|---|---|
| `Docker is running` | `docker version` cannot reach a daemon. Remedy: install and start Docker |
| `Docker Hub short description of <app> fits` | `short_description` is longer than 100 characters |
| `Docker registries of <app> agree` | `registry` and `registries` contradict each other |

```text
✖ Docker is running
    docker version could not reach a daemon
    → install and start Docker (https://docs.docker.com/get-docker/)
```

## Verified

smith 1.62.0, Windows 11. With `[conventions.docker] registry = "ghcr.io/acme"` and `tag_commit =
true`: scaffolding a Go app (the Dockerfile above), `smith explain svc:image`, `smith release
--dry-run`, `smith ci generate` (the `packages: write` permission and `GITHUB_TOKEN`) and `smith
doctor` (the daemon check failing). **Not verified:** a real `docker build` and push. Docker
Desktop was installed but its daemon was not running (`error during connect: ... dockerDesktopLinuxEngine`),
and pushing needs a registry login.

## See also

- [The node plugin](node.md) (`prune`), [go](go.md), [vite](vite.md)
- [Configuration](../../../reference/configuration.md): `[conventions.docker]`
- [ADR 0008: Containers through a docker plugin](../../../architecture/decisions/0008-containers-through-docker-plugin.md)
