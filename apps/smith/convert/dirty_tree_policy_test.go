package convert

import "testing"

func TestDirtyApartFromInstall(t *testing.T) {
	cases := []struct {
		name, status string
		want         bool
	}{
		{"clean", "", false},
		{"only what the install made", "?? .nx/\n?? package-lock.json\n", false},
		{"a tracked lockfile that changed is the person's", " M package-lock.json\n", true},
		{"another new file", "?? .nx/\n?? notes.txt\n", true},
		{"a change in a source file", " M src/a.ts\n", true},
	}
	for _, c := range cases {
		if got := DirtyApartFromInstall(c.status); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
