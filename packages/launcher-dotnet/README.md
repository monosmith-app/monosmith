# monosmith

The `smith` command for .NET users: `dotnet tool install --global monosmith`.

This tool is a launcher, not smith itself (ADR 0101). It runs the smith your workspace pins
(`smith = "<version>"` under `[workspace]` in smith.toml), the one `SMITH_VERSION` names, or the newest you
have. The first time it needs a version, it downloads smith's release archive, checks it against the
signed `checksums.txt`, and keeps it in `~/.smith/versions` (`SMITH_HOME` moves it). Behind a firewall,
`SMITH_DOWNLOAD_URL` points it at a mirror of the release archives.

See the [installation guide](https://github.com/monosmith-app/monosmith/blob/main/docs/getting-started/installing.md).
