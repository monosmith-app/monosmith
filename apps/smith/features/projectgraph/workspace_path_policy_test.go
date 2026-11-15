package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCleanRelative(t *testing.T) {
	for in, want := range map[string]string{
		`apps\web`:        "apps/web",
		"./libs//core/":   "libs/core",
		"libs/x/../Core":  "libs/Core",
		".":               ".",
		`Apps\Web\Src`:    "Apps/Web/Src",
		"packages/sdk/./": "packages/sdk",
	} {
		if got, err := CleanRelative(in); err != nil || got != want {
			t.Errorf("CleanRelative(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"/etc", `C:\work`, "c:/work", "..", "../outside", `libs\..\..\x`} {
		if _, err := CleanRelative(bad); err == nil {
			t.Errorf("CleanRelative(%q) must fail", bad)
		}
	}
}

func TestNormalisePaths(t *testing.T) {
	p := &Project{Project: protocol.Project{Name: "web", Root: `apps\web\`, Tasks: map[string]protocol.Task{
		"build": {Cwd: `.\apps\web`, Inputs: []string{`src\**\*.ts`, `!dist\**`}, Outputs: []string{`dist\`}},
	}}}
	if err := normalisePaths(p); err != nil {
		t.Fatal(err)
	}
	b := p.Tasks["build"]
	if p.Root != "apps/web" || b.Cwd != "apps/web" || strings.Join(b.Inputs, ",") != "src/**/*.ts,!dist/**" || b.Outputs[0] != "dist/" {
		t.Fatalf("%+v %+v", p.Root, b)
	}
	escaping := &Project{Project: protocol.Project{Name: "x", Root: "../x"}}
	if err := normalisePaths(escaping); err == nil || !strings.Contains(err.Error(), "leaves the workspace") {
		t.Fatalf("%v", err)
	}
}
