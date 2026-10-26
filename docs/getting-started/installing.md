# Installing smith

smith is one static binary per platform. Every channel below installs that
same binary as the `smith` command; none of them needs the language it comes
through at run time, and nothing is downloaded after install.

| Platform | Architectures | Minimum |
|---|---|---|
| Linux | x64, arm64 | any distribution (static, no libc needed) |
| macOS | Intel, Apple silicon | macOS 12 |
| Windows | x64, arm64 | Windows 10 |

## Channels

| Channel | Install | Upgrade | Uninstall |
|---|---|---|---|
| Install script (Linux, macOS, Git Bash) | `curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh \| sh` | run it again | delete `~/.local/bin/smith` |
| Install script (Windows PowerShell) | `irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1 \| iex` | run it again | delete `%USERPROFILE%\.local\bin\smith.exe` (and the directory from your user PATH) |
| npm | `npm install --global monosmith` | `npm install --global monosmith@latest` | `npm uninstall --global monosmith` |
| pipx | `pipx install monosmith` | `pipx upgrade monosmith` | `pipx uninstall monosmith` |
| uv | `uv tool install monosmith` | `uv tool upgrade monosmith` | `uv tool uninstall monosmith` |
| pip | `pip install monosmith` | `pip install --upgrade monosmith` | `pip uninstall monosmith` |
| .NET tool (.NET 10 SDK) | `dotnet tool install --global monosmith` | `dotnet tool update --global monosmith` | `dotnet tool uninstall --global monosmith` |
| RubyGems | `gem install monosmith` | `gem update monosmith` | `gem uninstall monosmith` |
| pub.dev | `dart pub global activate monosmith` | run it again | `dart pub global deactivate monosmith` |
| GitHub Releases | download `monosmith_<version>_<os>_<arch>.zip` from [releases](https://github.com/monosmith-app/monosmith/releases) and put `smith` on your PATH | download the newer zip | delete the binary |

To run smith once without installing it: `uvx --from monosmith smith --version`
or `npx monosmith --version`.

`@monosmith/cli` is the same npm package under the scope; `monosmith` is the
short name. Both install the binary through per-platform packages
(`@monosmith/cli-linux-x64`, …) that npm picks by `os` and `cpu`.

The pub.dev package is published by hand for now, so it can lag behind the
others.

### The install scripts

`install.sh` installs into `$SMITH_INSTALL_DIR` (default `~/.local/bin`);
`install.ps1` into `$env:SMITH_INSTALL_DIR` (default `$HOME\.local\bin`) and
adds that directory to your user PATH. Both take a version:

```sh
curl -fsSL https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.sh | sh -s -- 1.3.1
```

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/monosmith-app/monosmith/main/install.ps1))) -Version 1.3.1
```

`SMITH_VERSION` does the same. With no version they install the newest
`smith@` release. On GitHub Actions and Azure Pipelines `install.sh` also
adds its directory to `PATH` for the following steps; generated pipelines
install smith by running the project's own copy, `sh tools/smith/install.sh "<version>"`, pinned to the version that generated them (`smith ci generate` writes that copy and `smith doctor --fix` refreshes it; nothing is fetched from a branch at run time). Set
`GITHUB_TOKEN` to avoid GitHub's API rate limit on shared runners.

### Linux packages

Each GitHub Release also carries a `.deb` and an `.rpm` for x64 and arm64, with
the binary in `/usr/bin`:

```sh
sudo apt install ./monosmith_<version>_amd64.deb
sudo dnf install ./monosmith-<version>-1.x86_64.rpm
```

Upgrade by installing the newer file; uninstall with `sudo apt remove
monosmith` or `sudo dnf remove monosmith`.

### Not yet

Homebrew, Scoop, winget, Chocolatey, the AUR, Snap, an apt or dnf repository
and crates.io are tracked in #225. Each GitHub Release already carries a
Homebrew formula, a Scoop manifest and winget manifests, but no tap, bucket or
winget-pkgs submission uses them yet. `go install` is not supported, by
decision: smith is built from the repository's `go.work`, which `go install`
ignores ([ADR 0019](../architecture/decisions/0019-no-go-install.md)).

## Checking a download

Every release has a `checksums.txt` with the SHA-256 of each archive:

```sh
sha256sum --check --ignore-missing checksums.txt     # Linux
shasum -a 256 --check --ignore-missing checksums.txt # macOS
```

```powershell
(Get-FileHash monosmith_1.3.1_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
# compare with the line for that file in checksums.txt
```

The install scripts check the archive against `checksums.txt` before
installing. The npm packages carry
[provenance](https://docs.npmjs.com/generating-provenance-statements): npm
shows, and `npm audit signatures` checks, that they were built by this
repository's CI.

Archives, Linux packages and SBOMs are also attested with SLSA build
provenance on GitHub, which proves they were built by this repository's
workflow:

```sh
gh attestation verify monosmith_1.3.1_linux_amd64.tar.gz --repo monosmith-app/monosmith
```

Each release also carries a CycloneDX SBOM (`*_sbom.cdx.json`).

## Unsigned binaries

The binaries are not signed with an Apple or Microsoft certificate. That
matters only for a zip downloaded in a browser:

- **macOS** quarantines it, and Gatekeeper refuses to run the binary. Clear
  the flag once: `xattr -d com.apple.quarantine smith`.
- **Windows** SmartScreen warns about an unknown publisher the first time;
  choose *More info*, then *Run anyway*.

The install scripts and every package manager above download without marking
the file, so neither check applies to them.
