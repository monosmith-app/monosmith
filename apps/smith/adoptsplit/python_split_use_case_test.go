package adoptsplit

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const pyproject = `[project]
name = "orders"
dependencies = ["requests>=2.31", "uvicorn[standard]"]

[project.optional-dependencies]
dev = ["pytest>=8", "ruff>=0.5", "pytest-cov", "hypothesis>=6"]
docs = ["mkdocs"]

[tool.ruff]
line-length = 100

[tool.ruff.lint]
select = ["E", "F"]

[tool.mypy]
strict = true
`

func read(t *testing.T, root, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestSortPythonSeparatesToolsFromUnknown(t *testing.T) {
	root := t.TempDir()
	put(t, root, "pyproject.toml", pyproject)
	got, err := SortPython(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Sorted{Tools: []string{"pytest", "pytest-cov", "ruff"}, Unsure: []string{"hypothesis"}}); !reflect.DeepEqual(got, want) {
		t.Errorf("SortPython = %+v, want %+v", got, want)
	}
}

func TestSplitPythonMovesToolsAndRuffToTheWorkspace(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/pyproject.toml", pyproject)
	res, err := SplitPython(Request{Root: root, Dir: "apps/orders", Hoist: []string{"pytest", "ruff", "pytest-cov", "hypothesis"}})
	if err != nil {
		t.Fatal(err)
	}
	project := read(t, root, "apps/orders/pyproject.toml")
	for _, gone := range []string{"pytest", "hypothesis", "tool.ruff", "line-length", "[project.optional-dependencies]\n\n"} {
		if strings.Contains(project, gone) {
			t.Errorf("the project keeps %q:\n%s", gone, project)
		}
	}
	for _, kept := range []string{`dependencies = ["requests>=2.31", "uvicorn[standard]"]`, "[tool.mypy]", "strict = true"} {
		if !strings.Contains(project, kept) {
			t.Errorf("the project lost %q:\n%s", kept, project)
		}
	}
	if !strings.Contains(project, `docs = ["mkdocs"]`) {
		t.Errorf("the docs extra is not a development list and stays:\n%s", project)
	}
	reqs := read(t, root, "requirements-dev.txt")
	for _, want := range []string{"pytest==", "pytest-cov\n", "hypothesis>=6\n"} {
		if !strings.Contains(reqs, want) {
			t.Errorf("requirements-dev.txt lacks %q:\n%s", want, reqs)
		}
	}
	if strings.Count(reqs, "pytest==") != 1 || strings.Contains(reqs, "pytest>=8") || strings.Contains(reqs, "ruff>=0.5") {
		t.Errorf("a tool smith pins already is not added twice:\n%s", reqs)
	}
	ruff := read(t, root, "ruff.toml")
	if !strings.HasPrefix(ruff, "line-length = 100") || !strings.Contains(ruff, "[lint]\nselect = [\"E\", \"F\"]") || strings.Contains(ruff, "tool.") {
		t.Errorf("ruff.toml:\n%s", ruff)
	}
	if !strings.Contains(strings.Join(res.Notes, "\n"), "ruff.toml") {
		t.Errorf("notes = %v", res.Notes)
	}
}

func TestSplitPythonKeepsTheRuffTablesBesideARootRuffConfig(t *testing.T) {
	root := t.TempDir()
	put(t, root, "apps/orders/pyproject.toml", pyproject)
	put(t, root, "ruff.toml", "line-length = 88\n")
	if _, err := SplitPython(Request{Root: root, Dir: "apps/orders"}); err != nil {
		t.Fatal(err)
	}
	if read(t, root, "ruff.toml") != "line-length = 88\n" || !strings.Contains(read(t, root, "apps/orders/pyproject.toml"), "[tool.ruff]") {
		t.Error("the root ruff.toml wins and the project keeps its tables")
	}
}

func TestRequirementName(t *testing.T) {
	for in, want := range map[string]string{"pytest>=8": "pytest", "Pytest_Cov[x]>=1; python_version>'3'": "pytest-cov", "zope.interface": "zope-interface", "": "", "# comment": ""} {
		if got := RequirementName(in); got != want {
			t.Errorf("RequirementName(%q) = %q, want %q", in, got, want)
		}
	}
}
