# lazygit: install notes

## What was installed

- Version 0.66.2 (Windows, amd64), from the GitHub release `lazygit_0.66.2_windows_x86_64.zip`.
- Command: `gh release download -R jesseduffield/lazygit -p "*windows_x86_64.zip"`, then `unzip`.
- Location: `.tools/lazygit/lazygit.exe`. Nothing else was installed.
- Windows is supported natively. WSL was not needed.

## How it was run

- The sample repository is `.tools/samples/gitplay`. The script `.tools/mk-gitplay.sh` recreates it.
  It has 3 commits on `main`, 2 side branches (one conflicts with `main`), one stash, 2 modified files and 1 untracked file.
- Command: `lazygit.exe --use-config-dir .tools/lazygit-config` with the sample repository as the working directory.
- The driver is `.tools/drive.py` (my copy of `tui_drive.py`). It adds PNG output and mouse movement.
  `.tools/run.sh <tool> <cwd> <cmd> <steps> [cols rows]` wraps it.

## Things to know

- The first start of lazygit ran once against the real `monosmith` repository to test the driver. I sent no key except `Enter`.
  This run created lazygit state in its default user config folder. The later runs used `--use-config-dir`.
- Terminal size was 130x38 unless a file name says otherwise.
- Every `.png` is drawn by my script from the `pyte` screen model, not from a real GUI terminal. The `.txt` file next to it is the raw text capture.
