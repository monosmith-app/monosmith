package pluginpack

import "testing"

func TestTheIgnoreFileLeavesOutWhatItListsAndNothingElse(t *testing.T) {
	patterns := parseIgnore("# build output\r\n\r\nnode_modules/\n*.log\n/src\ndocs/**/*.draft.md\n")
	cases := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"node_modules", true, true},
		{"node_modules", false, false}, // "dir/" matches directories only
		{"deep/node_modules", true, true},
		{"run.log", false, true},
		{"deep/run.log", false, true},
		{"src", true, true},
		{"deep/src", true, false}, // a leading slash anchors at the root
		{"docs/a/b.draft.md", false, true},
		{"docs/b.md", false, false},
		{"bin/widget", false, false},
	}
	for _, c := range cases {
		if got := ignored(patterns, c.rel, c.isDir); got != c.want {
			t.Errorf("ignored(%q, dir=%v) = %v, want %v", c.rel, c.isDir, got, c.want)
		}
	}
}
