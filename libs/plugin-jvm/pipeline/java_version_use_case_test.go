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

func TestAJdkOlderThanCIHasASmithFixThatPinsAndRegenerates(t *testing.T) {
	f, _ := javaVersionFinding("openjdk version \"17.0.9\" 2023-10-17\n", "21")
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Files) != 1 || f.Fix.Files[0].Path != ".java-version" || f.Fix.Files[0].Content != "17\n" {
		t.Fatalf("%+v", f.Fix)
	}
	if len(f.Fix.Steps) != 1 || f.Fix.Steps[0].Argv[0] != "smith" || f.Fix.Steps[0].Argv[1] != "ci" {
		t.Fatalf("the step runs smith: %+v", f.Fix.Steps)
	}
}
