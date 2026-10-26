package venv

import (
	"fmt"
	"strings"
	"testing"
)

// pip-audit flagged setuptools 84.0.0 (GHSA-xmj3-9gf7-h52w, fixed in 84.1.0) in
// every workspace smith generated.
func TestTheSetuptoolsPinIsPastTheAdvisory(t *testing.T) {
	for _, line := range strings.Split(Requirements, "\n") {
		version, ok := strings.CutPrefix(line, "setuptools==")
		if !ok {
			continue
		}
		var major, minor int
		if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
			t.Fatal(err)
		}
		if major < 84 || (major == 84 && minor < 1) {
			t.Fatalf("setuptools==%s is older than the fixed 84.1.0", version)
		}
		return
	}
	t.Fatal("setuptools is not pinned")
}
