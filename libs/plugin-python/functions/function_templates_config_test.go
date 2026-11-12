package functions

import (
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"
)

// No host writes a script of smith's into the app: the tasks run smith's own
// commands and the tool of the host (decision D3).
func TestNoHostTemplateWritesAScript(t *testing.T) {
	for _, host := range []string{functionhost.Azure, functionhost.AWS, functionhost.GCP} {
		for name := range Templates(host) {
			if base := filepath.Base(name); base == "package.py" || base == "start.py" {
				t.Errorf("%s writes %s", host, name)
			}
		}
	}
}
