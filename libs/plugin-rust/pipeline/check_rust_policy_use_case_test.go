package pipeline

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeRust(t *testing.T, root, name, src string) {
	t.Helper()
	p := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRustPolicyFailsOnlyAboveTheBudgetAndSkipsTarget(t *testing.T) {
	root := t.TempDir()
	writeRust(t, root, "src/lib.rs", "fn a() { unsafe { } }\nfn b() { unsafe { } }\n")
	writeRust(t, root, "target/debug/gen.rs", "unsafe fn z() {}\n")
	if got, err := CheckRustPolicy([]string{root}, 2); err != nil || got != nil {
		t.Fatalf("within budget: %v %v", got, err)
	}
	got, err := CheckRustPolicy([]string{root}, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.ToSlash(filepath.Join(root, "src", "lib.rs")) + ": 2 unsafe", "2 unsafe in all, more than the 1 allowed (audit_max_unsafe)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%q", got)
	}
}
