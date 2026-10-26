package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
)

func TestCheckPythonPolicyDoesNotRequireHashesInTheGeneratedToolchainFile(t *testing.T) {
	if got := CheckPythonPolicy(venv.RequirementsFile, []byte(venv.Requirements), ""); len(got) != 0 {
		t.Fatalf("the generated requirements-dev.txt must pass: %v", got)
	}
}

func TestCheckPythonPolicyStillChecksTheToolchainFilesIndexAndUrls(t *testing.T) {
	src := "--extra-index-url https://evil.example/simple\nrepo @ git+https://example.com/r.git\n"
	want := []string{
		"requirements-dev.txt: extra index https://evil.example/simple: packages could come from outside the expected index",
		"requirements-dev.txt: repo is fetched from a URL or git, not from the index",
	}
	if got := CheckPythonPolicy(venv.RequirementsFile, []byte(src), ""); !reflect.DeepEqual(got, want) {
		t.Fatalf("%v", got)
	}
}

func TestCheckPythonPolicyStillRequiresHashesInAProjectsRequirements(t *testing.T) {
	if got := CheckPythonPolicy("apps/api/requirements.txt", []byte("flask==3.0.0\n"), ""); len(got) != 1 {
		t.Fatalf("%v", got)
	}
}
