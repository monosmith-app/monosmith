package updatenotice

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Blocker is another running smith that holds the installed binary.
type Blocker struct {
	// PID is the process id.
	PID int
	// Command is the command line, such as `smith.exe mcp`.
	Command string
}

// BlockingProcesses lists the other running smith programs that use the same
// binary as this one. Windows will not replace a binary that runs, so each of
// them (an editor's smith mcp, smith serve, another interface) stops an update.
// Other systems replace a running binary without trouble and report none.
//
// Parameters: exe is the path of the running binary; self is this process id;
// goos is runtime.GOOS.
// Returns: the other processes; none when the system cannot tell.
func BlockingProcesses(ctx context.Context, exe string, self int, goos string) []Blocker {
	if goos != "windows" || exe == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process -Filter "Name='smith.exe'" | ForEach-Object { "$($_.ProcessId)|$($_.ExecutablePath)|$($_.CommandLine)" }`).Output()
	if err != nil {
		return nil
	}
	return parseBlockers(string(out), exe, self)
}

// parseBlockers reads one "pid|path|command line" per line and keeps the
// processes that run exe, other than self.
func parseBlockers(out, exe string, self int) []Blocker {
	var found []Blocker
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "|", 3)
		if len(parts) < 3 {
			continue
		}
		pid, err := strconv.Atoi(parts[0])
		if err != nil || pid == self || !strings.EqualFold(parts[1], exe) {
			continue
		}
		found = append(found, Blocker{PID: pid, Command: parts[2]})
	}
	return found
}
