package declaredimports

import (
	"errors"
	"strings"
	"testing"
)

// fakeGo stands in for the go command: version is what `go version` prints,
// tidy answers `go mod tidy -diff`.
func fakeGo(t *testing.T, version string, tidyOut, tidyErr string, tidyCode int) *[][]string {
	t.Helper()
	var calls [][]string
	old := goCommand
	t.Cleanup(func() { goCommand = old })
	goCommand = func(dir string, env []string, args ...string) (string, string, int, error) {
		calls = append(calls, append(append([]string{}, env...), args...))
		if args[0] == "version" {
			if version == "" {
				return "", "", 0, errors.New("go: not found")
			}
			return version, "", 0, nil
		}
		return tidyOut, tidyErr, tidyCode, nil
	}
	return &calls
}

func plainModule(t *testing.T) string {
	root := t.TempDir()
	write(t, root, "m/go.mod", "module example.com/m\n\ngo 1.22\n\nrequire example.com/dep v1.0.0\n")
	write(t, root, "m/m.go", "package m\n\nimport _ \"example.com/dep\"\n")
	return root
}

func TestCheckAsksTidyForADiffAndSwitchesTheWorkspaceOff(t *testing.T) {
	calls := fakeGo(t, "go version go1.24.0 linux/amd64", "--- go.mod\n+++ go.mod\n@@\n-x\n", "", 1)
	result, err := Check(plainModule(t), "m")
	if err != nil {
		t.Fatal(err)
	}
	if result.Method != "go mod tidy -diff" || result.Clean() || !strings.Contains(result.Diff, "+++ go.mod") {
		t.Errorf("got %+v", result)
	}
	last := (*calls)[len(*calls)-1]
	if strings.Join(last, " ") != "GOWORK=off mod tidy -diff" {
		t.Errorf("tidy ran as %v", last)
	}
}

func TestCheckPassesWhenTidyIsClean(t *testing.T) {
	fakeGo(t, "go version go1.24.0 linux/amd64", "", "", 0)
	result, err := Check(plainModule(t), "m")
	if err != nil || !result.Clean() {
		t.Errorf("got %+v, %v", result, err)
	}
}

func TestCheckComparesTheSourcesWhenGoHasNoTidyDiff(t *testing.T) {
	calls := fakeGo(t, "go version go1.22.0 linux/amd64", "", "", 0)
	root := plainModule(t)
	write(t, root, "m/extra.go", "package m\n\nimport _ \"example.com/nowhere\"\n")
	result, err := Check(root, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(*calls) != 1 || !strings.Contains(result.Method, "no go mod tidy -diff") || len(result.Issues) != 1 {
		t.Errorf("calls %v, result %+v", *calls, result)
	}
}

func TestCheckFallsBackWhenTidyCannotRun(t *testing.T) {
	fakeGo(t, "go version go1.24.0 linux/amd64", "", "dial tcp: lookup proxy.golang.org: no such host\nmore", 1)
	result, err := Check(plainModule(t), "m")
	if err != nil || !result.Clean() || !strings.Contains(result.Method, "could not run: dial tcp") {
		t.Errorf("got %+v, %v", result, err)
	}
}

func TestCheckNeverAsksTidyAboutAModuleThatLeansOnGoWork(t *testing.T) {
	calls := fakeGo(t, "go version go1.24.0 linux/amd64", "", "", 1)
	root := t.TempDir()
	write(t, root, "go.work", "go 1.24\n\nuse (\n\t./a\n\t./b\n)\n")
	write(t, root, "a/go.mod", "module example.com/a\n\ngo 1.24\n")
	write(t, root, "a/a.go", "package a\n\nimport _ \"example.com/b\"\n")
	write(t, root, "b/go.mod", "module example.com/b\n\ngo 1.24\n")
	result, err := Check(root, "a")
	if err != nil || !result.Clean() || len(*calls) != 0 || !strings.Contains(result.Method, "go.work") {
		t.Errorf("calls %v, result %+v, %v", *calls, result, err)
	}
}
