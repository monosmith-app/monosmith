package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// npmLockfile is the part of package-lock.json (lockfileVersion 2 or 3) the
// policy reads.
type npmLockfile struct {
	Packages map[string]struct {
		Resolved         string `json:"resolved"`
		Integrity        string `json:"integrity"`
		Link             bool   `json:"link"`
		HasInstallScript bool   `json:"hasInstallScript"`
	} `json:"packages"`
}

// CheckNpmPolicy is the Node audit policy over a package-lock.json: every
// registry package carries an integrity hash, none is fetched from a git
// repository or over plain http, and only the packages in allowScripts run
// install scripts. It returns one line per violation, sorted, so the same
// lockfile always reads the same.
func CheckNpmPolicy(lockfile []byte, allowScripts []string) ([]string, error) {
	var lock npmLockfile
	if err := json.Unmarshal(lockfile, &lock); err != nil {
		return nil, fmt.Errorf("package-lock.json: %w", err)
	}
	if len(lock.Packages) == 0 {
		return nil, errors.New("package-lock.json has no packages table: use lockfileVersion 2 or 3")
	}
	var violations []string
	for key, p := range lock.Packages {
		const marker = "node_modules/"
		i := strings.LastIndex(key, marker)
		if key == "" || p.Link || i < 0 {
			continue // the root, a workspace link or member: not a dependency
		}
		name := key[i+len(marker):]
		switch {
		case strings.HasPrefix(p.Resolved, "git+"), strings.HasPrefix(p.Resolved, "git:"), strings.HasPrefix(p.Resolved, "github:"):
			violations = append(violations, fmt.Sprintf("%s is fetched from git (%s): depend on a published version", name, p.Resolved))
		case strings.HasPrefix(p.Resolved, "http://"):
			violations = append(violations, fmt.Sprintf("%s is fetched over plain http (%s)", name, p.Resolved))
		case strings.HasPrefix(p.Resolved, "https://") && p.Integrity == "":
			violations = append(violations, fmt.Sprintf("%s has no integrity hash in the lockfile", name))
		}
		if p.HasInstallScript && !slices.Contains(allowScripts, name) {
			violations = append(violations, fmt.Sprintf("%s runs an install script and is not in audit_allow_install_scripts", name))
		}
	}
	sort.Strings(violations)
	return violations, nil
}
