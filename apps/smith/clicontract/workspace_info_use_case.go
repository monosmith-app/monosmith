package clicontract

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/pelletier/go-toml/v2"
)

// releasesURL lists the GitHub Releases (by tag name, not by version); tests
// replace it.
var releasesURL = "https://api.github.com/repos/monosmith-app/monosmith/releases?per_page=100"

// WorkspaceReport says which smith runs, the newest one released, and the
// workspace under dir with its smith.toml.
func WorkspaceReport(version, dir string) Info {
	info := Info{CLI: CLIInfo{Version: version}}
	if latest := latestRelease(); latest != "" {
		info.CLI.Latest = &latest
		info.CLI.UpdateAvailable = newer(latest, version)
	}
	if root, err := workspace.FindRoot(dir); err == nil {
		ws := &WorkspaceInfo{Root: filepath.ToSlash(root), Config: map[string]any{}}
		if raw, err := os.ReadFile(filepath.Join(root, workspace.FileName)); err == nil {
			_ = toml.Unmarshal(raw, &ws.Config)
		}
		ws.Cache = cacheInfo(root)
		info.Workspace = ws
	}
	return info
}

// latestRelease is the version of the newest smith release, or "" when it
// cannot be found out quickly (offline, rate-limited, SMITH_NO_UPDATE_CHECK).
func latestRelease() string {
	if os.Getenv("SMITH_NO_UPDATE_CHECK") != "" {
		return ""
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(releasesURL)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	var releases []struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&releases) != nil {
		return ""
	}
	// GitHub lists releases by tag name as text (smith@0.9.0 before
	// smith@0.20.0), so the newest is the highest version, not the first.
	best := ""
	for _, r := range releases {
		if v, ok := strings.CutPrefix(r.Tag, "smith@"); ok && !r.Draft && !r.Prerelease && (best == "" || newer(v, best)) {
			best = v
		}
	}
	return best
}

// newer reports whether latest is a higher X.Y.Z than current; a development
// build is never behind.
func newer(latest, current string) bool {
	l, c := strings.Split(latest, "."), strings.Split(current, ".")
	if len(l) != 3 || len(c) != 3 {
		return false
	}
	for i := range l {
		if l[i] != c[i] {
			return atoi(l[i]) > atoi(c[i])
		}
	}
	return false
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}
		n = n*10 + int(r-'0')
	}
	return n
}
