package release

import "testing"

func TestFilesAreCountedByKind(t *testing.T) {
	for _, c := range []struct {
		names []string
		want  string
	}{
		{smithSample().Attached, "6 zips, 4 deb/rpm, checksums, SBOM, winget, brew, manifest"},
		{[]string{"a.zip"}, "1 zip"},
		{[]string{"a.tar.gz", "b.tgz", "notes.txt"}, "2 tarballs, 1 other file"},
		{nil, ""},
	} {
		if got := summariseFiles(c.names); got != c.want {
			t.Errorf("%v: %q, want %q", c.names, got, c.want)
		}
	}
}
