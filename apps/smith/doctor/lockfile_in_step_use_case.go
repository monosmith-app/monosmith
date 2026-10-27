package doctor

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lockfileCheck is the name of the check that package-lock.json matches
// package.json.
const lockfileCheck = "package-lock.json matches package.json"

// lockfileDryRunLimit bounds the dry run. It reads the lockfile and the
// manifests and installs nothing (about a second on 1,200 packages), so a run
// that takes longer is npm waiting on something, and says nothing about the
// lockfile.
const lockfileDryRunLimit = 30 * time.Second

// lockfileChecks runs npm ci --dry-run for an npm workspace, the way CI would
// install it.
func lockfileChecks(s *session.Session) []protocol.Finding {
	return checkLockfile(s.Root, dryRunNpmCi)
}

// checkLockfile asks run (npm ci --dry-run in root) whether package-lock.json
// matches package.json. It says nothing for a repository that is not an npm
// workspace with a lockfile, or when npm cannot answer.
func checkLockfile(root string, run func(root string) (output string, ran bool, err error)) []protocol.Finding {
	if _, err := os.Stat(filepath.Join(root, "package-lock.json")); err != nil || !hasWorkspaces(root) {
		return nil
	}
	output, ran, err := run(root)
	if !ran {
		return nil
	}
	f := protocol.Finding{Check: lockfileCheck, OK: err == nil, Warning: true}
	if err == nil {
		return []protocol.Finding{f}
	}
	f.Detail = "npm ci refuses package-lock.json, so a clean install (CI, a new clone) fails: " + firstNpmError(output)
	f.Remedy = "run npm install, then commit package-lock.json (`smith doctor --fix` does the install)"
	install := []string{"npm", "install", "--no-audit", "--no-fund"}
	f.Fix = &protocol.Fix{
		Description: "run npm install so package-lock.json matches package.json",
		Steps: []protocol.FixStep{
			{Description: "install the packages and refresh package-lock.json", Argv: install},
			{Description: "install again so the lockfile settles (npm can leave it one step behind)", Argv: install},
			{Description: "check that a clean install from the lockfile works", Argv: []string{"npm", "ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund"}},
		},
	}
	return []protocol.Finding{f}
}

func hasWorkspaces(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return false
	}
	var manifest struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	return json.Unmarshal(raw, &manifest) == nil && len(manifest.Workspaces) > 0 && string(manifest.Workspaces) != "null"
}

// dryRunNpmCi runs npm ci --dry-run; ran is false when npm is missing or the
// run timed out, and so says nothing about the lockfile.
func dryRunNpmCi(root string) (string, bool, error) {
	if _, err := exec.LookPath("npm"); err != nil {
		return "", false, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), lockfileDryRunLimit)
	defer cancel()
	cmd := exec.CommandContext(ctx, "npm", "ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return "", false, nil
	}
	return string(out), true, err
}

// firstNpmError is the first line npm reports as an error, or what it printed.
func firstNpmError(output string) string {
	var first string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "npm error") {
			if text := strings.TrimSpace(strings.TrimPrefix(line, "npm error")); text != "" {
				return text
			}
		} else if first == "" && line != "" {
			first = line
		}
	}
	return first
}
