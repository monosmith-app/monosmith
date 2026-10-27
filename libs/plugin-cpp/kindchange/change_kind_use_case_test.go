package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

const coreLists = `cmake_minimum_required(VERSION 3.25)

project(core VERSION 0.0.0 LANGUAGES CXX)
set(SMITH_KIND "lib")

add_library(core_lib src/a.cpp)
add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../util util)
include(../../cmake/flags.cmake)
`

var workspace = map[string]string{
	"CMakeLists.txt":            "cmake_minimum_required(VERSION 3.25)\nadd_subdirectory(libs/core)\nadd_subdirectory(apps/cli)\n",
	"libs/core/CMakeLists.txt":  coreLists,
	"libs/util/CMakeLists.txt":  "project(util)\nset(SMITH_KIND \"lib\")\n",
	"apps/cli/CMakeLists.txt":   "project(cli)\nset(SMITH_KIND \"app\")\nadd_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../libs/core core)\nadd_executable(cli src/main.cpp)\n",
	"libs/other/CMakeLists.txt": "project(other)\nadd_subdirectory(../core core)\n",
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "cpp"},
		{Name: "util", Root: "libs/util", Language: "cpp"},
		{Name: "cli", Root: "apps/cli", Language: "cpp"},
		{Name: "other", Root: "libs/other", Language: "cpp"},
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/CMakeLists.txt"]
	if !strings.Contains(core, `set(SMITH_KIND "package")`) {
		t.Fatalf("the kind changes:\n%s", core)
	}
	if !strings.Contains(core, "${CMAKE_CURRENT_SOURCE_DIR}/../../libs/util util") || !strings.Contains(core, "include(../../cmake/flags.cmake)") {
		t.Fatalf("its own paths reach the same places from the new directory:\n%s", core)
	}
	if !strings.Contains(got["CMakeLists.txt"], "add_subdirectory(packages/core)") || !strings.Contains(got["CMakeLists.txt"], "add_subdirectory(apps/cli)") {
		t.Fatalf("the root file follows:\n%s", got["CMakeLists.txt"])
	}
	if !strings.Contains(got["apps/cli/CMakeLists.txt"], "${CMAKE_CURRENT_SOURCE_DIR}/../../packages/core core") {
		t.Fatalf("dependents follow, with their variable:\n%s", got["apps/cli/CMakeLists.txt"])
	}
	if !strings.Contains(got["libs/other/CMakeLists.txt"], "add_subdirectory(../../packages/core core)") {
		t.Fatalf("a sibling follows:\n%s", got["libs/other/CMakeLists.txt"])
	}
	if _, touched := got["libs/util/CMakeLists.txt"]; touched {
		t.Fatal("a project that does not use core is not edited")
	}
}

func TestLibToPackageAndBackLeavesTheFileByteForByte(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/core/CMakeLists.txt": coreLists})
	project := protocol.Project{Name: "core", Root: "libs/core", Language: "cpp"}
	for _, to := range []protocol.ProjectKind{protocol.ProjectPackage, protocol.ProjectLib} {
		r, err := Change(dir, protocol.ChangeKindParams{Project: project, To: to, NewRoot: "libs/core"})
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range r.Files {
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(e.Path)), []byte(e.Content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "libs/core/CMakeLists.txt")); string(got) != coreLists {
		t.Fatalf("round trip changed the file:\n%s", got)
	}
}

func TestMissingSettingIsAddedUnderProject(t *testing.T) {
	dir := writeFiles(t, map[string]string{"libs/x/CMakeLists.txt": "cmake_minimum_required(VERSION 3.25)\nproject(x)\nadd_library(x a.cpp)\n"})
	r, err := Change(dir, protocol.ChangeKindParams{Project: protocol.Project{Name: "x", Root: "libs/x", Language: "cpp"}, To: protocol.ProjectLib, NewRoot: "libs/x"})
	if err != nil {
		t.Fatal(err)
	}
	if got := edited(r)["libs/x/CMakeLists.txt"]; got != "cmake_minimum_required(VERSION 3.25)\nproject(x)\nset(SMITH_KIND \"lib\")\nadd_library(x a.cpp)\n" {
		t.Fatalf("got %q", got)
	}
}

func TestAppToLibKeepsTheProgramAndSaysSo(t *testing.T) {
	dir := writeFiles(t, workspace)
	r, err := Change(dir, protocol.ChangeKindParams{Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/cli", Projects: projects()})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "program") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
}
