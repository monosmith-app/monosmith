# monosmith

The `smith` command for Node.js users: `npm install --global monosmith`, or `npx monosmith`.

This package is a launcher, not smith itself (ADR 0101). It runs the smith your workspace pins
(`smith = "<version>"` under `[workspace]` in smith.toml), the one `SMITH_VERSION` names, or the newest you
have. The first time it needs a version, it downloads smith's release archive, checks the minisign
signature of `checksums.txt` and the archive's SHA-256, and keeps it in `~/.smith/versions`
(`SMITH_HOME` moves it). The download happens on the first run, not in an install script, so
`npm install --ignore-scripts` works. Behind a firewall, `SMITH_DOWNLOAD_URL` points it at a mirror of
the release archives. `smith update` moves to a newer smith without npm.

Up to 2.x this package was smith itself; those versions stay on npm but get no new smith.

See the [installation guide](https://github.com/monosmith-app/monosmith/blob/main/docs/getting-started/installing.md).
