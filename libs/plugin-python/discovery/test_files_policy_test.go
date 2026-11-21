package discovery

import "testing"

func TestIsTestFile(t *testing.T) {
	for name, want := range map[string]bool{
		"test_app.py": true, "app_test.py": true, "test_.py": true,
		"app.py": false, "testing.py": false, "test_app.pyc": false, "conftest.py": false, "contest.py": false,
	} {
		if got := IsTestFile(name); got != want {
			t.Errorf("IsTestFile(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestHasTestFiles(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "analyzer.py", "print(1)\n")
	write(t, dir, ".venv/lib/test_vendored.py", "")
	write(t, dir, "build/lib/test_old.py", "")
	if HasTestFiles(dir) {
		t.Fatal("files in .venv and build/ are not the project's tests")
	}
	write(t, dir, "tests/unit/test_analyzer.py", "def test_x(): pass\n")
	if !HasTestFiles(dir) {
		t.Fatal("tests/unit/test_analyzer.py is a test file")
	}
}

func TestDiscoverGivesNoTestTaskWithoutTests(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/tool/pyproject.toml", "[project]\nname = \"tool\"\nversion = \"0.1.0\"\n\n[tool.smith]\nkind = \"app\"\n")
	write(t, ws, "apps/tool/tool.py", "print(1)\n")
	write(t, ws, "apps/kept/pyproject.toml", "[project]\nname = \"kept\"\nversion = \"0.1.0\"\n\n[tool.pytest.ini_options]\npython_files = [\"check_*.py\"]\n")
	write(t, ws, "apps/kept/check_kept.py", "def test_x(): pass\n")
	write(t, ws, "libs/core/pyproject.toml", "[project]\nname = \"core\"\nversion = \"0.1.0\"\n")
	write(t, ws, "libs/core/core_test.py", "def test_x(): pass\n")

	res, err := Discover(ws, []string{"apps/kept/pyproject.toml", "apps/tool/pyproject.toml", "libs/core/pyproject.toml"})
	if err != nil {
		t.Fatal(err)
	}
	has := map[string]bool{}
	for _, p := range res.Projects {
		_, has[p.Name] = p.Tasks["test"]
		if _, ok := p.Tasks["lint"]; !ok {
			t.Errorf("%s keeps its lint task", p.Name)
		}
	}
	if has["tool"] {
		t.Error("a project with no test file gets no test task")
	}
	if !has["kept"] {
		t.Error("a project that configures pytest itself keeps its test task")
	}
	if !has["core"] {
		t.Error("a project with a *_test.py file keeps its test task")
	}
}
