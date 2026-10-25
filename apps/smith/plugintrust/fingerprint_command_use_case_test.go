package plugintrust

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func digest(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestMeasureHashesTheProgramAndAScriptItRuns(t *testing.T) {
	dir := t.TempDir()
	interpreter := writeFile(t, dir, "interp", "interpreter")
	script := writeFile(t, dir, "plugin.js", "script")
	fp, err := Measure([]string{interpreter, "--flag", script, "not-a-file"})
	if err != nil {
		t.Fatal(err)
	}
	if len(fp.Executables) != 2 {
		t.Fatalf("got %+v", fp.Executables)
	}
	if fp.Executables[0].SHA256 != digest("interpreter") || fp.Executables[1].SHA256 != digest("script") {
		t.Fatalf("digests %+v", fp.Executables)
	}
	if PinnedDigest(fp) != digest("script") {
		t.Fatalf("the pin should be the script's digest")
	}
}

func TestMeasureReportsAMissingProgram(t *testing.T) {
	if _, err := Measure([]string{filepath.Join(t.TempDir(), "nope")}); err == nil {
		t.Fatal("a missing program was measured")
	}
	if _, err := Measure(nil); err == nil {
		t.Fatal("an empty command was measured")
	}
}
