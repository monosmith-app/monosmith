package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func cppWorkspace(t *testing.T, files map[string]string) string {
	t.Helper()
	ws := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func engine(tasks map[string]protocol.Task) []protocol.Project {
	return []protocol.Project{{Name: "engine", Root: "libs/engine", Tasks: tasks}}
}

var smithTasks = discovery.Tasks(discovery.Shape{Root: "libs/engine", Kind: protocol.ProjectLib})

const onePreset = "{\n  \"version\": 3,\n  \"configurePresets\": [\n    {\n      \"name\": \"dev\",\n      \"binaryDir\": \"${sourceDir}/build\"\n    }\n  ]\n}\n"

func TestCMakePresetThatBuildsIntoBuildMovesToTheProjectsDistFolder(t *testing.T) {
	ws := cppWorkspace(t, map[string]string{"libs/engine/CMakePresets.json": onePreset})
	findings, ok := CheckBuildDir(ws, engine(smithTasks))
	if !ok || len(findings) != 1 || findings[0].OK || findings[0].Warning {
		t.Fatalf("%+v", findings)
	}
	f := findings[0]
	if !strings.Contains(f.Detail, "libs/engine/CMakePresets.json builds into libs/engine/build/, not into dist/libs/engine/") {
		t.Fatalf("where it builds now and where it should: %s", f.Detail)
	}
	if f.Fix == nil || len(f.Fix.Choices[0].Options) != 2 || len(f.Fix.Choices[0].Options[1].Fix.Files) != 0 {
		t.Fatalf("a question, move or leave: %+v", f.Fix)
	}
	got := f.Fix.Choices[0].Options[0].Fix.Files[0]
	want := strings.Replace(onePreset, "${sourceDir}/build", "${sourceDir}/../../dist/libs/engine", 1)
	if got.Path != "libs/engine/CMakePresets.json" || got.Content != want {
		t.Fatalf("only the value changes: %+v", got)
	}
}

func TestCMakeSeveralPresetsEachGetTheirOwnTree(t *testing.T) {
	presets := "{\"configurePresets\":[{\"name\":\"a\",\"binaryDir\":\"${sourceDir}/build/a\"},{\"name\":\"b\",\"binaryDir\":\"build/b\"}]}"
	findings, _ := CheckBuildDir(cppWorkspace(t, map[string]string{"libs/engine/CMakePresets.json": presets}), engine(smithTasks))
	got := findings[0].Fix.Choices[0].Options[0].Fix.Files[0].Content
	want := "{\"configurePresets\":[{\"name\":\"a\",\"binaryDir\":\"${sourceDir}/../../dist/libs/engine/${presetName}\"},{\"name\":\"b\",\"binaryDir\":\"${sourceDir}/../../dist/libs/engine/${presetName}\"}]}"
	if got != want {
		t.Fatalf("%s", got)
	}
}

func TestCMakePassesWhenTheTaskAndThePresetsBuildCentrally(t *testing.T) {
	presets := strings.Replace(onePreset, "${sourceDir}/build", "${sourceDir}/../../dist/libs/engine", 1)
	findings, _ := CheckBuildDir(cppWorkspace(t, map[string]string{"libs/engine/CMakePresets.json": presets}), engine(smithTasks))
	if len(findings) != 1 || !findings[0].OK {
		t.Fatalf("%+v", findings)
	}
}

func TestCMakeHandMadeTaskWithAnotherTreeIsReportedWithoutAFix(t *testing.T) {
	tasks := map[string]protocol.Task{"configure": {Command: []string{"cmake", "-S", ".", "-B", "build"}}}
	findings, _ := CheckBuildDir(cppWorkspace(t, nil), engine(tasks))
	f := findings[0]
	if f.OK || f.Fix != nil || !strings.Contains(f.Detail, "its configure task builds into libs/engine/build/") || !strings.Contains(f.Remedy, "-B dist/libs/engine") {
		t.Fatalf("a command line written by hand is not rewritten: %+v", f)
	}
}

func TestCMakeRefusesWithAWarningAndTheReason(t *testing.T) {
	for name, tc := range map[string]struct {
		files map[string]string
		want  string
	}{
		"a Dockerfile":                          {map[string]string{"libs/engine/CMakePresets.json": onePreset, "libs/engine/Dockerfile": "COPY build/engine /usr/bin/engine\n"}, "libs/engine/Dockerfile names libs/engine/build/"},
		"a sibling project that reads the tree": {map[string]string{"libs/engine/CMakePresets.json": onePreset, "libs/other/CMakeLists.txt": "include(../engine/build/gen.cmake)\n"}, "libs/other/CMakeLists.txt reads a file of engine"},
	} {
		t.Run(name, func(t *testing.T) {
			projects := append(engine(smithTasks), protocol.Project{Name: "other", Root: "libs/other"})
			findings, _ := CheckBuildDir(cppWorkspace(t, tc.files), projects)
			f := findings[0]
			if f.OK || !f.Warning || f.Fix != nil || !strings.Contains(f.Remedy, tc.want) {
				t.Fatalf("a warning with the reason and no fix: %+v", f)
			}
		})
	}
}

func TestCMakeLeavesAVariableBinaryDirUnchecked(t *testing.T) {
	presets := "{\"configurePresets\":[{\"name\":\"a\",\"binaryDir\":\"$env{BUILD_ROOT}/a\"}]}"
	if f, ok := CheckBuildDir(cppWorkspace(t, map[string]string{"libs/engine/CMakePresets.json": presets}), engine(nil)); ok {
		t.Fatalf("nothing is guessed: %+v", f)
	}
}
