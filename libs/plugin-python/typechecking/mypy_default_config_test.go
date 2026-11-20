package typechecking

import (
	"regexp"
	"strings"
	"testing"
)

func TestTheDefaultConfigExcludesTheSdistStagingFolder(t *testing.T) {
	var pattern string
	for _, line := range strings.Split(DefaultConfig(), "\n") {
		if v, ok := strings.CutPrefix(line, "exclude = "); ok {
			pattern = v
		}
	}
	re, err := regexp.Compile(pattern)
	if err != nil || pattern == "" {
		t.Fatalf("exclude %q: %v", pattern, err)
	}
	for path, want := range map[string]bool{
		"monosmith-0.0.0/src/monosmith_launcher/x.py": true,
		"libs/core/core-1.2/src/core/a.py":            true,
		"src/monosmith_launcher/x.py":                 false,
		"scripts/migrate-2024.py":                     false,
	} {
		if got := re.MatchString(path); got != want {
			t.Errorf("%s: got %v, want %v", path, got, want)
		}
	}
}
