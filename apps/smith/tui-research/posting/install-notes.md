# posting install notes

Version run: Posting 2.11.1 (Python, Textual). Platform: Windows 11, native. WSL not used.

## Install

`PIPX_HOME=.tools/pipx PIPX_BIN_DIR=.tools/bin pipx install posting` (pipx 1.17.6, Python 3.12). The executable is `.tools/bin/posting.exe`.
`posting --version` is not an option. The version shows in the top bar.

## Sample collection

`.tools/samples/pc/` holds five `*.posting.yaml` requests (`Hello local`, `Create user`, `Slow call`, `Server error`, `Unreachable`). A small Python server `.tools/samples/srv.py` on `127.0.0.1:8765` serves `/ok`, `/users` (POST), `/slow` (5 seconds), `/boom` (HTTP 500). Port 9 gives a refused connection.
Command: `posting.exe --collection <path to pc>`, 130 columns by 38 rows.

## Driver

`.tools/drive.py` (shared, ConPTY plus pyte, colour PNG per shot). Mouse clicks work through SGR codes. Ctrl keys were sent as raw control bytes (`\x0a` for Ctrl+J, `\x10` for Ctrl+P, `\x0f` for Ctrl+O). F1 was sent as `\x1bOP`.

## Problems seen

Some clicks hit the wrong row because the list moves after a selection. Delete from the collection list was tried with Backspace and Delete. A confirmation dialog was not captured. Not verified whether Posting asks before it deletes.
