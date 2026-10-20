package pipeline

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckPythonPolicyAcceptsHashedPinsFromTheExpectedIndex(t *testing.T) {
	src := "# tools\n--index-url https://pypi.org/simple/\nrequests==2.32.0 \\\n    --hash=sha256:aa \\\n    --hash=sha256:bb\nblack==24.1.0 --hash=sha256:cc # fmt\n-r other.txt\n"
	if got := CheckPythonPolicy("r.txt", []byte(src), ""); len(got) != 0 {
		t.Errorf("%v", got)
	}
}

func TestCheckPythonPolicyReportsEachViolation(t *testing.T) {
	src := "requests==2.32.0\n--extra-index-url https://evil.example/simple\nlib @ https://example.com/lib.whl\ngit+https://github.com/a/b.git#egg=b\n-i https://other.example/simple\n"
	want := []string{
		"r.txt: requests==2.32.0 has no --hash",
		"r.txt: extra index https://evil.example/simple: packages could come from outside the expected index",
		"r.txt: lib is fetched from a URL or git, not from the index",
		"r.txt: git+https://github.com/a/b.git#egg=b is fetched from a URL or git, not from the index",
		"r.txt: index https://other.example/simple is not the expected index https://pypi.org/simple",
	}
	if got := CheckPythonPolicy("r.txt", []byte(src), ""); !reflect.DeepEqual(got, want) {
		t.Errorf("%q", got)
	}
}

func TestCheckPythonPolicyUsesTheWorkspacesPrivateIndex(t *testing.T) {
	feed := "https://pkgs.example/simple/"
	if got := CheckPythonPolicy("r.txt", []byte("--index-url "+feed+"\n"), feed); len(got) != 0 {
		t.Errorf("%v", got)
	}
	if got := CheckPythonPolicy("r.txt", []byte("--index-url https://pypi.org/simple\n"), feed); len(got) != 1 {
		t.Errorf("%v", got)
	}
}

func TestCheckPythonPolicyFilesSkipsMissingFiles(t *testing.T) {
	dir := t.TempDir()
	present := filepath.Join(dir, "requirements.txt")
	if err := os.WriteFile(present, []byte("flask==3.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := CheckPythonPolicyFiles([]string{filepath.Join(dir, "none.txt"), present}, "")
	if err != nil || len(got) != 1 {
		t.Errorf("%v %v", got, err)
	}
}
