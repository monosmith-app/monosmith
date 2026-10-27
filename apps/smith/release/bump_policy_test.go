package release

import (
	"testing"
)

func TestParseCommit(t *testing.T) {
	cases := []struct {
		subject, body string
		typ           string
		scopes        int
		breaking      bool
		valid         bool
	}{
		{"feat(web): add login", "", "feat", 1, false, true},
		{"fix(web,ui)!: rename prop", "", "fix", 2, true, true},
		{"refactor: tidy", "BREAKING CHANGE: removed x", "refactor", 0, true, true},
		{"chore(deps): bump", "BREAKING-CHANGE: y", "chore", 1, true, true},
		{"Merge pull request #1", "", "", 0, false, false},
		{"feat:missing space", "", "", 0, false, false},
	}
	for _, c := range cases {
		got := ParseCommit("h", c.subject, c.body)
		if got.Type != c.typ || len(got.Scopes) != c.scopes || got.Breaking != c.breaking || got.Valid != c.valid {
			t.Errorf("%q: %+v", c.subject, got)
		}
	}
}

func TestLevels(t *testing.T) {
	level := func(subject string) Level { return LevelOf(ParseCommit("", subject, "")) }
	if level("feat: x") != Minor || level("fix: x") != Patch || level("perf: x") != Patch || level("feat!: x") != Major {
		t.Fatal("feat/fix/perf/breaking")
	}
	for _, s := range []string{"docs: x", "chore: x", "test: x", "ci: x", "refactor: x", "build: x", "style: x", "chore(release): publish", "random"} {
		if level(s) != None {
			t.Errorf("%q must release nothing", s)
		}
	}
	if LevelOfAll([]Commit{ParseCommit("", "fix: a", ""), ParseCommit("", "feat: b", ""), ParseCommit("", "docs: c", "")}) != Minor {
		t.Fatal("highest level wins")
	}
}

func TestBump(t *testing.T) {
	cases := map[string][2]string{
		"0.0.0 minor":      {"0.0.0", "0.1.0"},
		"0.0.0 patch":      {"0.0.0", "0.0.1"},
		"0.3.1 major":      {"0.3.1", "0.4.0"},
		"1.2.3 major":      {"1.2.3", "2.0.0"},
		"1.2.3 minor":      {"1.2.3", "1.3.0"},
		"1.2.3 patch":      {"1.2.3", "1.2.4"},
		"1.3.0-rc.1 patch": {"1.3.0-rc.1", "1.3.0"},
	}
	levels := map[string]Level{"major": Major, "minor": Minor, "patch": Patch}
	for name, c := range cases {
		v, err := ParseVersion(c[0])
		if err != nil {
			t.Fatal(err)
		}
		var l Level
		for k, lv := range levels {
			if len(name) > len(k) && name[len(name)-len(k):] == k {
				l = lv
			}
		}
		if got := v.Bump(l).String(); got != c[1] {
			t.Errorf("%s: got %s want %s", name, got, c[1])
		}
	}
	a, _ := ParseVersion("1.2.3")
	b, _ := ParseVersion("v1.10.0")
	pre, _ := ParseVersion("1.2.3-beta")
	if !a.Less(b) || b.Less(a) || !pre.Less(a) {
		t.Fatal("ordering")
	}
	if _, err := ParseVersion("1.2"); err == nil {
		t.Fatal("1.2 is not semver")
	}
}

func TestDirective(t *testing.T) {
	d, ok := ParseDirective("version(web)[2.0.0]: start 2.x")
	if !ok || d.Project != "web" || d.Version != "2.0.0" {
		t.Fatalf("%+v %v", d, ok)
	}
	if _, ok := ParseDirective("feat(web): x"); ok {
		t.Fatal("not a directive")
	}
}

func TestUncommittedIgnoresNpmrcWrittenByTheNpmSignIn(t *testing.T) {
	if got := Uncommitted(" M .npmrc\n M apps/web/.npmrc\n"); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}

func TestUncommittedIgnoresTheWorkspaceChecksumCache(t *testing.T) {
	got := Uncommitted(" M go.work.sum\n M apps/web/package.json\nM  libs/x/go.mod\n")
	if len(got) != 2 || got[0] != "apps/web/package.json" || got[1] != "libs/x/go.mod" {
		t.Fatalf("%v", got)
	}
	if got := Uncommitted(" M go.work.sum\n"); len(got) != 0 {
		t.Fatalf("%v", got)
	}
}
