package convertplan

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestAPackageTaskGetsItsOwnDirectoryUnderDistDrop(t *testing.T) {
	o := workspace.TaskOverride{
		Command: []string{"node", "-e", "fs.mkdirSync('dist/drop',{recursive:true});z.writeZip('dist/drop/node-function-app-azure-host.zip')"},
		Outputs: []string{"../../dist/drop"},
	}
	dir, ok := ownDropDir("apps/azure-host", &o)
	if !ok || dir != "dist/drop/azure-host" {
		t.Fatalf("%q %v", dir, ok)
	}
	want := []string{"node", "-e", "fs.mkdirSync('dist/drop/azure-host',{recursive:true});z.writeZip('dist/drop/azure-host/node-function-app-azure-host.zip')"}
	if !reflect.DeepEqual(o.Command, want) {
		t.Errorf("command %q", o.Command)
	}
	if !reflect.DeepEqual(o.Outputs, []string{"../../dist/drop/azure-host"}) {
		t.Errorf("outputs %q", o.Outputs)
	}
}

func TestOnlyTheSharedDropIsMoved(t *testing.T) {
	o := workspace.TaskOverride{Command: []string{"zip", "dist/drop/x.zip"}, Outputs: []string{"../../dist/apps/web"}}
	if _, ok := ownDropDir("apps/web", &o); ok || o.Command[1] != "dist/drop/x.zip" {
		t.Errorf("an own build directory stays as it is: %+v", o)
	}
	o = workspace.TaskOverride{Command: []string{"zip", "dist/drop/web/x.zip"}, Outputs: []string{"../../dist/drop/web"}}
	if _, ok := ownDropDir("apps/web", &o); ok {
		t.Errorf("an own drop directory stays as it is: %+v", o)
	}
}

func TestMoveDropPathsLeavesLongerNamesAndOwnPathsAlone(t *testing.T) {
	for in, want := range map[string]string{
		"dist/drop":                       "dist/drop/web",
		"{workspaceRoot}/dist/drop/a.zip": "{workspaceRoot}/dist/drop/web/a.zip",
		"dist/drop/web/a.zip":             "dist/drop/web/a.zip",
		"dist/drop-old/a.zip":             "dist/drop-old/a.zip",
		"dist/dropped":                    "dist/dropped",
		"cp dist/drop/a dist/drop/b":      "cp dist/drop/web/a dist/drop/web/b",
		"no mention":                      "no mention",
		"'dist/drop'":                     "'dist/drop/web'",
	} {
		if got := moveDropPaths(in, "dist/drop/web"); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
