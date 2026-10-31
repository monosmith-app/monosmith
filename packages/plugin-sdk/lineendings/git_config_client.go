package lineendings

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	settingsMu sync.Mutex
	settings   = map[string]string{}
)

// gitSetting returns the value git has for key in the repository at root, or
// "" when it is unset, when root is not inside a work tree (no checkout
// converts line endings there) or when git cannot run. The answer is kept for
// the process.
func gitSetting(root, key string) string {
	cacheKey := root + "\x00" + key
	settingsMu.Lock()
	defer settingsMu.Unlock()
	if v, ok := settings[cacheKey]; ok {
		return v
	}
	value := ""
	if exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree").Run() == nil {
		if out, err := exec.Command("git", "-C", root, "config", "--get", key).Output(); err == nil {
			value = strings.ToLower(strings.TrimSpace(string(out)))
		}
	}
	settings[cacheKey] = value
	return value
}
