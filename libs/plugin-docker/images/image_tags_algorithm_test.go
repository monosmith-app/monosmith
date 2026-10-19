package images

import (
	"reflect"
	"strings"
	"testing"
)

func TestTagsAreLatestAndTheVersionByDefault(t *testing.T) {
	got, missing := Tags(Target{}, "1.2.3", "0123456789abcdef", "main")
	if !reflect.DeepEqual(got, []string{"latest", "1.2.3"}) || len(missing) != 0 {
		t.Errorf("%v %v", got, missing)
	}
}

func TestTheCommitAndBranchTagsComeBeforeLatestAndTheVersionIsLast(t *testing.T) {
	got, missing := Tags(Target{TagCommit: true, TagBranch: true}, "1.2.3", "0123456789ABCDEF", "release/2.x")
	want := []string{"sha-0123456", "release-2.x", "latest", "1.2.3"}
	if !reflect.DeepEqual(got, want) || len(missing) != 0 {
		t.Errorf("%v %v", got, missing)
	}
}

func TestATagTheReleaseCannotNameIsLeftOutAndSaid(t *testing.T) {
	got, missing := Tags(Target{TagCommit: true, TagBranch: true}, "1.0.0", "", "")
	if !reflect.DeepEqual(got, []string{"latest", "1.0.0"}) || len(missing) != 2 {
		t.Errorf("%v %v", got, missing)
	}
	if got, _ := Tags(Target{TagCommit: true}, "1.0.0", "abc", ""); len(got) != 2 {
		t.Errorf("a hash shorter than %d characters is no commit: %v", CommitTagLength, got)
	}
}

func TestBranchTagIsAValidImageTag(t *testing.T) {
	for in, want := range map[string]string{
		"main":                   "main",
		"refs/heads/main":        "main",
		"feature/Add Thing!":     "feature-Add-Thing-",
		"-leading":               "leading",
		".hidden":                "hidden",
		"v1.0_rc-2":              "v1.0_rc-2",
		"///":                    "",
		strings.Repeat("a", 200): strings.Repeat("a", 128),
	} {
		if got := BranchTag(in); got != want {
			t.Errorf("%q → %q, want %q", in, got, want)
		}
	}
}
