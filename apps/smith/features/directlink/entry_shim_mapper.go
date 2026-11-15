package directlink

import "strings"

// shimText is the text of an entry that calls the program at target. A batch
// file and a PowerShell script have a form of their own, and a shell script
// calls the program with exec, so no process is left in front of it.
func shimText(kind Kind, target string) string {
	switch kind {
	case CmdShim:
		// The goto ends the batch context before the program starts, as npm's own
		// shim does: a batch file is read line by line while it runs, so an update that
		// writes the file again must not have its new lines run when the program ends.
		return "@ECHO off\r\nendLocal & goto #_undefined_# 2>NUL || title %COMSPEC% & \"" + target + "\" %*\r\n"
	case Ps1Shim:
		return "#!/usr/bin/env pwsh\r\n& \"" + target + "\" $args\r\nexit $LASTEXITCODE\r\n"
	default:
		// Forward slashes: a shell on Windows reads C:/dir, not C:\dir.
		return "#!/bin/sh\nexec \"" + strings.ReplaceAll(target, "\\", "/") + "\" \"$@\"\n"
	}
}
