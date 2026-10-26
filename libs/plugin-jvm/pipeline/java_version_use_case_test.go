package pipeline

import "testing"

func TestJavaVersionFinding(t *testing.T) {
	for _, c := range []struct {
		out      string
		found    bool
		ok, warn bool
	}{
		{"openjdk version \"21.0.5\" 2024-10-15\n", true, true, false},
		{"openjdk version \"17.0.9\" 2023-10-17\n", true, false, false},
		{"java version \"25.0.1\" 2025-10-21\n", true, false, true},
		{"java version \"1.8.0_392\"\n", true, false, false},
		{"", false, false, false},
	} {
		f, found := javaVersionFinding(c.out, "21")
		if found != c.found || f.OK != c.ok || f.Warning != c.warn {
			t.Errorf("%q: %v %+v", c.out, found, f)
		}
	}
}
