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
	write("apps/cli/cli-1.2.0-1.rockspec", `-- smithKind: app
package = "cli"
version = "1.2.0-1"
dependencies = {
  "lua >= 5.1",
  "strings ~> 2.0",
  "penlight",
}
`)
	write("apps/cli/spec/cli_spec.lua", "")
	write("apps/cli/cli-scm-1.rockspec", "package = \"cli\"\n")
	write("libs/strings/strings-2.0.0-3.rockspec", "-- smithStack: tools\npackage = \"strings\"\nversion = \"2.0.0-3\"\n")
	write("libs/odd/odd.rockspec", "-- no package\n")
	res, err := Discover(ws, []string{"root.rockspec", "apps/cli/cli-1.2.0-1.rockspec", "apps/cli/cli-scm-1.rockspec", "libs/strings/strings-2.0.0-3.rockspec", "libs/odd/odd.rockspec"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	cli, strs := res.Projects[0], res.Projects[1]
	if cli.Name != "cli" || cli.Kind != protocol.ProjectApp || cli.Version != "1.2.0" || strings.Join(cli.Imports, ",") != "strings,penlight" || cli.Exports[0] != "cli" {
		t.Errorf("cli %+v", cli)
	}
	if got := strings.Join(cli.Tasks["build"].Command, " "); got != "luarocks make --tree ../../dist/apps/cli cli-1.2.0-1.rockspec" || cli.Tasks["build"].Outputs[0] != "../../dist/apps/cli" {
		t.Errorf("build %q", got)
	}
	if _, ok := cli.Tasks["test"]; !ok {
		t.Error("a project with spec/ has a test task")
	}
	if strs.Kind != protocol.ProjectLib || strs.Stack != "tools" || strs.Version != "2.0.0" {
		t.Errorf("strings %+v", strs)
	}
	if _, ok := strs.Tasks["test"]; ok {
		t.Error("a project without spec/ has no test task")
	}
}

func TestReleaseVersion(t *testing.T) {
	for in, want := range map[string]string{"1.2.0-1": "1.2.0", "scm-1": "scm", "1.0": "1.0", "": ""} {
		if got := ReleaseVersion(in); got != want {
			t.Errorf("ReleaseVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
