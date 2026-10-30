# Release files of a private repository, used by another repository's CI

A private app repository (call it `acme/app`) releases files, and a second
repository (`acme/site`, say) needs them in its own pipeline: to bundle them,
to deploy them, to test against them. The second repository's `GITHUB_TOKEN`
reaches only itself, so it cannot see the first one's release. A fine-grained
personal access token can, with one read permission.

## What smith does, and what you set up

smith, in `acme/app`: the release job tags the app and creates its GitHub
Release. The files come from either a plugin's archives or the `build` command
of a `[projects.<app>.release]` table
([Release files from a command](../reference/releasing.md#release-files-from-a-command)),
and are attached with the workflow's own `GITHUB_TOKEN`
(`contents: write`, which the generated workflow already has). Nothing in
`acme/app` needs the token below.

You, once: create the token, store it as a secret in `acme/site`, and download
with it. smith does not generate the download step in `acme/site`, even when that
repository is smith-managed; add it as a step of your own (the `before-setup`
slot of a generated workflow keeps hand-written steps through
`smith ci generate`).

## The token

1. In GitHub, go to Settings, Developer settings, Personal access tokens,
   Fine-grained tokens, Generate new token. Sign in as an account that can read
   `acme/app`; the token has that account's access, so a machine user or a
   service account is better than a person.
2. Resource owner: the organisation or user that owns `acme/app`. An
   organisation that requires approval for fine-grained tokens must approve it
   (Organization settings, Personal access tokens, Pending requests).
3. Repository access: Only select repositories, then `acme/app` alone.
4. Permissions, Repository permissions: Contents, Read-only. Metadata,
   Read-only is added for you. Nothing else; a release download needs no write
   and no Actions, Packages or Pull requests permission.
5. Set an expiry (90 days, say) and put the renewal in a calendar: the pipeline
   fails with a 404 when the token expires, because GitHub answers 404 for a
   private repository it will not show.
6. In `acme/site`: Settings, Secrets and variables, Actions, New repository
   secret, named `APP_RELEASE_TOKEN`, with the token as its value.

An organisation with a GitHub App can use that instead: install it on
`acme/app` with Contents read-only and mint a short-lived token in the workflow
with `actions/create-github-app-token`, which needs no renewal.

## The download in `acme/site`

```yaml
      - name: Download the app's release files
        env:
          GH_TOKEN: ${{ secrets.APP_RELEASE_TOKEN }}
        run: |
          gh release download "app@1.4.0" --repo acme/app --pattern '*.zip' --pattern '*.zip.sha256' --dir vendor/app
```

The tag is `<project>@<version>` unless `tag_prefix` or `bare_tags` says
otherwise (`smith release --status` lists what was released). To take the latest
release of the app, list tags with `gh release list --repo acme/app` or pin the
version in a file the pipeline reads; do not use `releases/latest` of a
monorepo, which is the newest release of any project.

`gh` is installed on GitHub-hosted runners. Without it, ask the API for the
asset by its id, because the `browser_download_url` of a private release needs
a browser session:

```sh
curl -fsSL -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/octet-stream" \
  "https://api.github.com/repos/acme/app/releases/assets/$ASSET_ID" -o app.zip
```

## Checking it works

- `404 Not Found` or `Could not resolve to a Repository`: the token cannot see
  `acme/app`. Check the resource owner, that the repository is selected, that the
  organisation approved the token and that it has not expired.
- `403 Resource not accessible by personal access token`: the Contents
  permission is missing or is not read-only at least.
- `no assets match the file patterns`: the release exists but the build has not
  uploaded yet (the `assets` job of the app's workflow runs after the release
  job), or the pattern does not match the file names.

Verify the files you download against the `.sha256` file the release carries,
when its build writes one, before you use them.
