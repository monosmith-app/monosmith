# btop install notes

Tool run: btop4win v1.0.5 (the Windows port of btop++). Platform: Windows 11, native. WSL not used.

## Install

1. `gh release download v1.0.5 -R aristocratos/btop4win -p btop4win-x64.zip` into `.tools/dl-btop/`.
2. Unzip. The binary is `.tools/dl-btop/btop4win/btop4win.exe`. A `btop.conf` and a `themes/` folder sit beside it.
3. Command used: `btop4win.exe`, 130 columns by 40 rows. It runs with no flags.

## Driver

`.tools/drive.py` (shared, ConPTY plus pyte, colour PNG per shot). Screens are PNG renders of the pyte screen, plus `.txt` dumps.

## Problems seen

- A bare `Esc` key is slow to act. Wait 1.5 seconds after `Esc`. Some key steps in the first run arrived out of order. Those shots are in `first-run-mixed/`. Treat them as unreliable.
- btop quit by itself in two runs after a key sequence that included `Esc` then `e`. Cause not found.
- Safety incident: during the first run the agent sent `k` while a process was selected. btop has a kill key. The agent cannot prove that no process died. Some `python` processes seen at the start were not seen at the end. They may have ended by themselves. `vmmemWSL` was still alive.
- The screens show real process names, user name and an IP address of this PC. Do not publish them.
