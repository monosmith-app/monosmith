package pluginarchive

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteIsDeterministicAndIndependentOfInputOrder(t *testing.T) {
	a := []Entry{memEntry("plugin.json", "{}", false), memEntry("bin/w", "binary", true), memEntry("z.txt", "z", false)}
	b := []Entry{a[2], a[0], a[1]}
	var one, two bytes.Buffer
	if _, err := Write(&one, a); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(&two, b); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(one.Bytes(), two.Bytes()) {
		t.Fatal("the same files packed to different bytes")
	}
}

func TestWriteRecordsEveryFileInTheChecksumsAndRefusesItsOwnName(t *testing.T) {
	var buf bytes.Buffer
	sums, err := Write(&buf, []Entry{memEntry("plugin.json", "{}", false), memEntry("bin/w", "binary", true)})
	if err != nil {
		t.Fatal(err)
	}
	if sums["bin/w"] != sum("binary") || sums["plugin.json"] != sum("{}") {
		t.Fatalf("got %v", sums)
	}
	entries, err := ZipEntries(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	last := entries[len(entries)-1]
	if last.Name != ChecksumsFile {
		t.Fatalf("checksums are not last: %s", last.Name)
	}
	if _, err := Write(&bytes.Buffer{}, []Entry{memEntry(ChecksumsFile, "x", false)}); err == nil || !strings.Contains(err.Error(), "written by smith") {
		t.Fatalf("got %v", err)
	}
	if _, err := Write(&bytes.Buffer{}, []Entry{memEntry("../x", "x", false)}); err == nil {
		t.Fatal("packed an unsafe name")
	}
}
