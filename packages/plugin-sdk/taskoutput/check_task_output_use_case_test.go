package taskoutput

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var spec = Spec{
	Match:   func(c []string) bool { return len(c) > 0 && c[0] == "tool" },
	Flags:   []string{"-o", "--outdir"},
	Flag:    "-o",
	Default: "beside the sources",
	Value:   func(t string) string { return t },
}

func project(command ...string) protocol.Project {
	return protocol.Project{Name: "kit", Root: "libs/kit", Tasks: map[string]protocol.Task{"build": {Command: command}}}
}

func TestCheckIsSilentWhenTheOptionNamesDist(t *testing.T) {
	for _, c := range [][]string{
		{"tool", "-o", "{workspaceRoot}/dist/libs/kit/bin"},
		{"tool", "--outdir={workspaceRoot}/dist/libs/kit"},
		{"tool", "-o", "{workspaceRoot}/dist/x/{target}"}, // another placeholder: not read
		{"other", "build"}, // not that command
	} {
		if f, ok := Check(project(c...), "build", spec); ok {
			t.Errorf("%v: %+v", c, f)
		}
	}
	if _, ok := Check(protocol.Project{Name: "kit"}, "build", spec); ok {
		t.Error("no build task")
	}
}

func TestCheckReportsACommandThatWritesElsewhereAndSaysWhatToPass(t *testing.T) {
	f, ok := Check(project("tool", "."), "build", spec)
	if !ok || f.OK || f.Check != "kit builds into dist/libs/kit" || !strings.Contains(f.Detail, "passes no -o") || !strings.Contains(f.Detail, "beside the sources") ||
		!strings.Contains(f.Remedy, "pass -o {workspaceRoot}/dist/libs/kit in the command of the build task of [projects.kit.tasks.build] in smith.toml") {
		t.Fatalf("%+v", f)
	}
	f, ok = Check(project("tool", "-o", "out/bin"), "build", spec)
	if !ok || !strings.Contains(f.Detail, "writes into libs/kit/out/bin, not into dist/libs/kit/") {
		t.Fatalf("a relative option is read from the project: %+v", f)
	}
	if f, ok = Check(project("tool", "--outdir=../../out"), "build", spec); !ok || !strings.Contains(f.Detail, "writes into out,") {
		t.Fatalf("%+v", f)
	}
}

func TestCheckAcceptsAWorkFolderUnderDistAndReadsTheDeclaredOutputs(t *testing.T) {
	p := project("tool", "-o", "{workspaceRoot}/dist/tool/libs/kit")
	if f, ok := Check(p, "build", spec); ok {
		t.Errorf("a work folder under dist/<tool> is accepted: %+v", f)
	}
	p.Tasks["build"] = protocol.Task{Command: []string{"tool", "-o", "{workspaceRoot}/dist/libs/kit/bin"}, Outputs: []string{"../../dist/libs/kit"}}
	if f, ok := Check(p, "build", spec); ok {
		t.Errorf("the output holds the build: %+v", f)
	}
	p.Tasks["build"] = protocol.Task{Command: []string{"tool", "-o", "{workspaceRoot}/dist/libs/kit"}, Outputs: []string{"../../dist/other"}}
	f, ok := Check(p, "build", spec)
	if !ok || f.OK || f.Warning || f.Problem != "caches dist/other, not where it builds" || !strings.Contains(f.Detail, "a cache hit restores nothing") ||
		f.Remedy != "list ../../dist/libs/kit in the outputs of [projects.kit.tasks.build] in smith.toml" {
		t.Fatalf("%+v", f)
	}
}

func TestCheckNamesTheProjectApartFromTheProblem(t *testing.T) {
	f, _ := Check(project("tool", "."), "build", spec)
	if f.Project != "kit" || f.Area != "" || strings.HasPrefix(f.Problem, "kit") || !strings.HasPrefix(f.Problem, "builds ") {
		t.Fatalf("%+v", f)
	}
}
