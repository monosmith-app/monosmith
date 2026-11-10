# yazi install notes

Version run: Yazi 26.9.1 (x86_64-pc-windows-msvc). Platform: Windows 11, native. WSL not used.

## Install

1. `gh release download v26.9.1 -R sxyazi/yazi -p yazi-x86_64-pc-windows-msvc.zip` into `.tools/dl-yazi/`.
2. Unzip. Binary: `.tools/dl-yazi/yazi-x86_64-pc-windows-msvc/yazi.exe`. `ya.exe` sits beside it.
3. Yazi needs the `file` command for MIME types. Set `YAZI_FILE_ONE=C:\Program Files\Git\usr\bin\file.exe` (from Git for Windows).
4. Start in a sandbox folder: `yazi.exe C:/Users/.../.tools/samples/yz`. A path with backslashes was read wrongly (Yazi opened `C:\`). Use forward slashes.

## Sandbox

`.tools/samples/yz` holds `docs/`, `src/`, `img/`, small files, `big.bin` (600 MB) and `huge.bin` (1.5 GB). The first aborted run started in `C:\` by mistake. It only opened a prompt and a trash confirmation that the agent cancelled with `n`. Nothing was deleted. Its screens are in `aborted-first-run/`.
A copy `yazi_1.cmd` and `big_1.bin` were created by test copies inside `.tools/samples`.

## Driver

`.tools/drive.py` (shared, ConPTY plus pyte, colour PNG per shot), 120 columns by 34 rows.
Nerd Font file icons show as empty boxes in the PNG because the render font has no such glyphs. The real terminal shows them.
