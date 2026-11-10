package clicontract

import "testing"

func TestTitleUsesTheNamePeopleUse(t *testing.T) {
	for in, want := range map[string]string{"cpp": "C++", "dotnet": ".NET", "go": "Go", "python": "Python", "applescript": "AppleScript"} {
		if got := title(in); got != want {
			t.Errorf("title(%q) = %q, want %q", in, got, want)
		}
	}
}
