package pluginarchive

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func memEntry(name, body string, exec bool) Entry {
	return Entry{Name: name, Executable: exec, Size: int64(len(body)), Open: func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}}
}

func sum(body string) string {
	h := sha256.Sum256([]byte(body))
	return hex.EncodeToString(h[:])
}

func packed(t *testing.T, entries ...Entry) []Entry {
	t.Helper()
	var buf bytes.Buffer
	if _, err := Write(&buf, entries); err != nil {
		t.Fatal(err)
	}
	out, err := ZipEntries(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestExtractRoundTripKeepsContentAndExecuteBit(t *testing.T) {
	entries := packed(t,
		memEntry("plugin.json", `{"name":"w"}`, false),
		memEntry("bin/w", "binary", true),
	)
	dest := t.TempDir()
	ex, err := Extract(entries, dest, Limits{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if ex.Files != 2 || !ex.Checksummed {
		t.Fatalf("got %+v", ex)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "bin", "w")); string(got) != "binary" {
		t.Fatalf("content %q", got)
	}
	if _, err := os.Stat(filepath.Join(dest, ChecksumsFile)); err != nil {
		t.Fatal("the checksums file is not kept")
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(filepath.Join(dest, "bin", "w"))
		if info.Mode()&0o111 == 0 {
			t.Fatal("lost the execute bit")
		}
		info, _ = os.Stat(filepath.Join(dest, "plugin.json"))
		if info.Mode()&0o111 != 0 {
			t.Fatal("gained an execute bit")
		}
	}
}

func TestExtractRefusesATamperedFile(t *testing.T) {
	good := packed(t, memEntry("plugin.json", "{}", false), memEntry("bin/w", "binary", true))
	tampered := []Entry{memEntry("bin/w", "evil!!", true)}
	for _, e := range good {
		if e.Name != "bin/w" {
			tampered = append(tampered, e)
		}
	}
	dest := t.TempDir()
	_, err := Extract(tampered, dest, Limits{}, true)
	if err == nil || !strings.Contains(err.Error(), "digest mismatch for bin/w") {
		t.Fatalf("got %v", err)
	}
}

func TestExtractRefusesFilesTheChecksumsDoNotList(t *testing.T) {
	good := packed(t, memEntry("plugin.json", "{}", false))
	extra := append(good, memEntry("extra.sh", "x", false))
	if _, err := Extract(extra, t.TempDir(), Limits{}, true); err == nil || !strings.Contains(err.Error(), "extra.sh is in the archive but not in") {
		t.Fatalf("unlisted file: %v", err)
	}
	var withoutManifest []Entry
	for _, e := range good {
		if e.Name != "plugin.json" {
			withoutManifest = append(withoutManifest, e)
		}
	}
	if _, err := Extract(withoutManifest, t.TempDir(), Limits{}, true); err == nil || !strings.Contains(err.Error(), "no plugin.json") {
		t.Fatalf("missing manifest: %v", err)
	}
}

func TestExtractNeedsChecksumsForAnArchiveButNotForADirectory(t *testing.T) {
	entries := []Entry{memEntry("plugin.json", "{}", false)}
	if _, err := Extract(entries, t.TempDir(), Limits{}, true); err == nil || !strings.Contains(err.Error(), ChecksumsFile) {
		t.Fatalf("got %v", err)
	}
	ex, err := Extract(entries, t.TempDir(), Limits{}, false)
	if err != nil || ex.Checksummed {
		t.Fatalf("got %+v, %v", ex, err)
	}
}

func TestExtractRefusesPathTraversalAndOddNames(t *testing.T) {
	for _, name := range []string{"../evil", "a/../../evil", "/etc/passwd", `..\evil`, "C:evil", "a//b", "./a", "nul", "dir/COM1.txt", "trailing."} {
		entries := []Entry{memEntry("plugin.json", "{}", false), memEntry(name, "x", false)}
		dest := t.TempDir()
		if _, err := Extract(entries, dest, Limits{}, false); err == nil || !strings.Contains(err.Error(), "unsafe entry name") {
			t.Errorf("%q: got %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(dest), "evil")); err == nil {
			t.Fatalf("%q escaped", name)
		}
	}
}

func TestZipEntriesRefusesSymbolicLinks(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: "link", Method: zip.Store}
	h.SetMode(0o777 | os.ModeSymlink)
	w, _ := zw.CreateHeader(h)
	_, _ = w.Write([]byte("/etc/passwd"))
	_ = zw.Close()
	if _, err := ZipEntries(bytes.NewReader(buf.Bytes()), int64(buf.Len())); err == nil || !strings.Contains(err.Error(), "link or special file") {
		t.Fatalf("got %v", err)
	}
}

func TestExtractRefusesNamesThatDifferOnlyInCase(t *testing.T) {
	entries := []Entry{memEntry("plugin.json", "{}", false), memEntry("A", "1", false), memEntry("a", "2", false)}
	if _, err := Extract(entries, t.TempDir(), Limits{}, false); err == nil || !strings.Contains(err.Error(), "differ only in case") {
		t.Fatalf("got %v", err)
	}
}

func TestExtractEnforcesSizeLimitsEvenWhenTheArchiveLiesAboutSize(t *testing.T) {
	big := strings.Repeat("a", 4096)
	declaredSmall := Entry{Name: "plugin.json", Size: 2, Open: func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(big)), nil
	}}
	if _, err := Extract([]Entry{declaredSmall}, t.TempDir(), Limits{MaxFileBytes: 1024}, false); err == nil || !strings.Contains(err.Error(), "larger than the size limit") {
		t.Fatalf("a file that lies about its size: %v", err)
	}
	if _, err := Extract([]Entry{memEntry("plugin.json", big, false)}, t.TempDir(), Limits{MaxFileBytes: 1024}, false); err == nil || !strings.Contains(err.Error(), "limit for one file") {
		t.Fatalf("declared too big: %v", err)
	}
	two := []Entry{memEntry("plugin.json", big, false), memEntry("b", big, false)}
	if _, err := Extract(two, t.TempDir(), Limits{MaxTotalBytes: 6000}, false); err == nil || !strings.Contains(err.Error(), "expands to more than") {
		t.Fatalf("total: %v", err)
	}
	if _, err := Extract(two, t.TempDir(), Limits{MaxFiles: 1}, false); err == nil || !strings.Contains(err.Error(), "limit is 1") {
		t.Fatalf("count: %v", err)
	}
}
