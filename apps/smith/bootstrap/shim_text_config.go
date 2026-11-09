package bootstrap

// ShellPath and CmdPath are where the launcher is written, workspace-relative.
const (
	ShellPath = "tools/smith/smith"
	CmdPath   = "tools/smith/smith.cmd"
)

// shellShim is the sh launcher. Order: the pin (else any smith on PATH); a
// smith on PATH that already is the pin; a copy cached under .smith/bin/<pin>;
// the committed installer (tools/smith/install.sh, written by smith ci
// generate); npx. A smith on PATH that is this launcher is skipped.
const shellShim = `#!/bin/sh
# Runs the smith version this workspace pins in smith.toml ([ci] smith_version),
# so a laptop and CI use the same smith. Written by smith bootstrap; the pin is
# read each run, so changing it needs no regeneration. Without a pin, or with
# "latest", it runs the smith on PATH.
set -eu
here=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
root=$(CDPATH='' cd -- "$here/../.." && pwd)
pin=$(sed -n 's/^[[:space:]]*smith_version[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$root/smith.toml" 2>/dev/null | head -n 1)

# on_path prints the first smith on PATH that is not this launcher.
on_path() {
  IFS=:; for d in $PATH; do
    [ -n "$d" ] || d=.
    for f in "$d/smith" "$d/smith.exe"; do
      if [ -f "$f" ] && [ -x "$f" ] && ! [ "$f" -ef "$here/smith" ]; then printf '%s\n' "$f"; return 0; fi
    done
  done
  return 1
}

installed=$(on_path || true)
if [ -z "$pin" ] || [ "$pin" = latest ]; then
  [ -n "$installed" ] || { echo "smith is not installed. Install it: npm i -g monosmith" >&2; exit 127; }
  exec "$installed" "$@"
fi
if [ -n "$installed" ] && [ "$("$installed" --version 2>/dev/null | sed 's/.* //')" = "$pin" ]; then
  exec "$installed" "$@"
fi
dir="$root/.smith/bin/$pin"
for f in "$dir/smith" "$dir/smith.exe"; do
  [ -x "$f" ] && exec "$f" "$@"
done
if [ -f "$here/install.sh" ]; then
  SMITH_INSTALL_DIR="$dir" sh "$here/install.sh" "$pin" >&2
  for f in "$dir/smith" "$dir/smith.exe"; do
    [ -x "$f" ] && exec "$f" "$@"
  done
fi
if command -v npx >/dev/null 2>&1; then
  exec npx --yes "monosmith@$pin" "$@"
fi
echo "smith $pin is pinned in smith.toml but not installed. Install it: npm i -g monosmith@$pin" >&2
exit 127
`

// cmdShim is the Windows launcher, with the same order as shellShim except
// that it has no installer step: the installer is a shell script.
const cmdShim = `@echo off
rem Runs the smith version this workspace pins in smith.toml ([ci] smith_version),
rem so a laptop and CI use the same smith. Written by smith bootstrap; the pin is
rem read each run. Without a pin, or with "latest", it runs the smith on PATH.
setlocal EnableDelayedExpansion
set "ROOT=%~dp0..\.."
set "PIN="
if exist "%ROOT%\smith.toml" for /f "tokens=2 delims==" %%a in ('findstr /b /c:"smith_version" "%ROOT%\smith.toml"') do (
  set "RAW=%%a"
  for %%b in (!RAW!) do if not defined PIN set "PIN=%%~b"
)
set "FOUND="
for /f "delims=" %%p in ('where smith 2^>nul') do if not defined FOUND if /i not "%%~dpp"=="%~dp0" set "FOUND=%%p"
if not defined PIN goto :onpath
if /i "!PIN!"=="latest" goto :onpath
if defined FOUND for /f "tokens=3" %%v in ('"!FOUND!" --version 2^>nul') do if "%%v"=="!PIN!" goto :onpath
set "BIN=%ROOT%\.smith\bin\!PIN!\smith.exe"
if exist "!BIN!" "!BIN!" %* & exit /b !errorlevel!
where npx >nul 2>nul
if not errorlevel 1 call npx --yes monosmith@!PIN! %* & exit /b !errorlevel!
echo smith !PIN! is pinned in smith.toml but not installed. Install it: npm i -g monosmith@!PIN! 1>&2
exit /b 127
:onpath
if not defined FOUND echo smith is not installed. Install it: npm i -g monosmith 1>&2 & exit /b 127
"!FOUND!" %*
exit /b !errorlevel!
`

// ShellShim returns the text of tools/smith/smith (LF line ends).
func ShellShim() string { return shellShim }

// CmdShim returns the text of tools/smith/smith.cmd (CRLF line ends, which cmd.exe needs).
func CmdShim() string { return crlf(cmdShim) }
