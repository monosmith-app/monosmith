package updatenotice

import "testing"

func TestParseBlockersKeepsOtherProcessesOfTheSameBinary(t *testing.T) {
	exe := `C:\npm\node_modules\monosmith\node_modules\@monosmith\cli-win32-x64\bin\smith.exe`
	out := "100|" + exe + "|smith.exe mcp\r\n" +
		"200|" + exe + "|smith.exe update\r\n" +
		`300|c:\other\node_modules\@monosmith\cli-win32-x64\bin\smith.exe|smith.exe serve` + "\r\n" +
		"garbage\r\n" +
		"400|" + exe + "|smith.exe serve\r\n"
	got := parseBlockers(out, exe, 200)
	if len(got) != 2 || got[0].PID != 100 || got[0].Command != "smith.exe mcp" || got[1].PID != 400 {
		t.Fatalf("blockers = %+v", got)
	}
}

func TestBlockingProcessesOnlyLooksOnWindows(t *testing.T) {
	if got := BlockingProcesses(t.Context(), "/usr/bin/smith", 1, "linux"); got != nil {
		t.Errorf("blockers = %v", got)
	}
}
