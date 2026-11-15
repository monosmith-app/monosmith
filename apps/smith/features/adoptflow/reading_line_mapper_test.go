package adoptflow

import "testing"

func TestReadingLineCountsTheFiles(t *testing.T) {
	cases := map[int]string{
		0:       "… reading the repository",
		1:       "… reading 1 file",
		23:      "… reading 23 files",
		1000:    "… reading 1,000 files",
		1234567: "… reading 1,234,567 files",
	}
	for n, want := range cases {
		if got := ReadingLine(n); got != want {
			t.Errorf("ReadingLine(%d) = %q, want %q", n, got, want)
		}
	}
}
