package adoptflow

import (
	"strings"
	"testing"
)

func TestAFileThatChangesInManyFoldersIsOneLine(t *testing.T) {
	c := Changes{
		Created: []string{"smith.toml"},
		Edited:  []string{"libs/a/package.json", "libs/b/package.json", "tsconfig.base.json"},
		Removed: []string{"libs/a/rollup.config.js", "libs/b/rollup.config.js", "nx.json"},
	}
	text := c.Text()
	for _, want := range []string{"create 1 files, edit 3 and remove 3", "package.json in 2 folders", "rollup.config.js in 2 folders", "• nx.json", "• tsconfig.base.json"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if c.Total() != 7 {
		t.Fatalf("total %d", c.Total())
	}
}

func TestChangesNeedThePreparedCopy(t *testing.T) {
	if _, err := (&Session{}).Changes(); err != ErrNotConvertible {
		t.Fatalf("err %v", err)
	}
}
