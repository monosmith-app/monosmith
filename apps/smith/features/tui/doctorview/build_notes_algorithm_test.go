package doctorview

import "testing"

func TestBuildMetadataLeavesTheCauseAndGoesToTheNotes(t *testing.T) {
	in := `C:\go\bin\golangci-lint.exe (golangci-lint has version 2.1.6 built with go1.24.2 from (unknown, modified: ?, mod sum: "h1:abcDEF0123456789abcdef=") on windows/amd64), older than the 2.5 CI installs`
	cause, notes := SplitBuildNotes(in)
	if want := `C:\go\bin\golangci-lint.exe (golangci-lint has version 2.1.6), older than the 2.5 CI installs`; cause != want {
		t.Errorf("cause %q, want %q", cause, want)
	}
	for _, part := range []string{"built with go1.24.2", "from (unknown, modified: ?", "on windows/amd64"} {
		if !contains(notes, part) {
			t.Errorf("notes %q lack %q", notes, part)
		}
	}
}

func TestAnUnknownPlatformIsMetadataToo(t *testing.T) {
	cause, notes := SplitBuildNotes("found /bin/x (x has version 2.14.0 built with go1.27.0 on (unknown))")
	if cause != "found /bin/x (x has version 2.14.0)" || !contains(notes, "on (unknown)") {
		t.Errorf("%q, %q", cause, notes)
	}
}

func TestASentenceWithoutMetadataIsLeftAsItIsOnOneLine(t *testing.T) {
	cause, notes := SplitBuildNotes("node is 18.0.0, CI uses 22\nsecond line")
	if cause != "node is 18.0.0, CI uses 22" || notes != "" {
		t.Errorf("%q, %q", cause, notes)
	}
}

func contains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
