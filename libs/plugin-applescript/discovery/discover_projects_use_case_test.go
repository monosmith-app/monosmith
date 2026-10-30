package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("apps/notes/main.applescript", "-- smithKind: app\n-- smithVersion: 1.2.0\non run\nend run\n")
	write("apps/notes/helper.applescript", "-- smithKind: lib\n")
	write("libs/dialogs/alpha.applescript", "-- smithStack: mac\n-- smithVersion: 0.3.0\n")
	write("libs/dialogs/beta.applescript", "")
	write("top.applescript", "")
	res, err := Discover(ws, []string{"top.applescript", "apps/notes/helper.applescript", "apps/notes/main.applescript", "libs/dialogs/alpha.applescript", "libs/dialogs/beta.applescript"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	notes, dialogs := res.Projects[0], res.Projects[1]
	if notes.Name != "notes" || notes.Kind != protocol.ProjectApp || notes.Version != "1.2.0" || notes.Exports[0] != "apps/notes" || len(notes.Imports) != 0 || notes.Tags[0] != NativeTag {
		t.Errorf("notes %+v", notes)
	}
	if got := strings.Join(notes.Tasks["build"].Command, " "); !strings.HasSuffix(got, "sh ../../dist/apps/notes notes.scpt main.applescript") || notes.Tasks["build"].Outputs[0] != "../../dist/apps/notes" {
		t.Errorf("build %q", got)
	}
	if dialogs.Kind != protocol.ProjectLib || dialogs.Stack != "mac" || dialogs.Version != "0.3.0" || !strings.HasSuffix(strings.Join(dialogs.Tasks["build"].Command, " "), "alpha.applescript") {
		t.Errorf("without a main.applescript the first script is the main one: %+v", dialogs)
	}
	if len(notes.Tasks) != 1 {
		t.Error("osacompile is the only tool: a build task and nothing else")
	}
}

func TestPickMain(t *testing.T) {
	for _, tc := range []struct {
		names []string
		want  string
	}{
		{[]string{"b.applescript", "main.applescript", "a.applescript"}, "main.applescript"},
		{[]string{"b.applescript", "a.applescript", "notes.txt"}, "a.applescript"},
		{[]string{"x.scpt"}, ""},
	} {
		if got := PickMain(tc.names); got != tc.want {
			t.Errorf("PickMain(%v) = %q, want %q", tc.names, got, tc.want)
		}
	}
}
