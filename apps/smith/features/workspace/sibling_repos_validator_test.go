package workspace

import (
	"strings"
	"testing"
)

func TestSiblingReposDefaultTheirDir(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"w\"\n[[ci.sibling_repos]]\nrepo = \"acme/VideoPlugins\"\nenv = \"VIDEOPLUGINS\"\npre_build = \"cargo build --release\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	r := cfg.CI.SiblingRepos[0]
	if r.Dir != "VideoPlugins" || r.Env != "VIDEOPLUGINS" || r.PreBuild != "cargo build --release" {
		t.Errorf("got %+v", r)
	}
}

func TestSiblingReposAreRefusedWithAFix(t *testing.T) {
	for name, tc := range map[string]struct{ toml, want string }{
		"not owner/name":   {"[[ci.sibling_repos]]\nrepo = \"VideoPlugins\"\nenv = \"X\"\n", "is not owner/name"},
		"bad env":          {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"1X\"\n", "environment variable name"},
		"no env":           {"[[ci.sibling_repos]]\nrepo = \"a/b\"\n", "environment variable name"},
		"dir escapes":      {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\ndir = \"../b\"\n", "plain directory"},
		"dir twice":        {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\n[[ci.sibling_repos]]\nrepo = \"c/b\"\nenv = \"Y\"\n", "used by another"},
		"env twice":        {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\n[[ci.sibling_repos]]\nrepo = \"a/c\"\nenv = \"X\"\n", "used by another"},
		"other provider":   {"[ci]\nproviders = [\"github\", \"gitlab\"]\n[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\n", "GitHub Actions only"},
		"ref whitespace":   {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\nref = \"a b\"\n", "holds whitespace"},
		"bad token secret": {"[[ci.sibling_repos]]\nrepo = \"a/b\"\nenv = \"X\"\ntoken_secret = \"a-b\"\n", "not a secret name"},
	} {
		_, err := Parse([]byte("[workspace]\nname = \"w\"\n" + tc.toml))
		if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "→") {
			t.Errorf("%s: err = %v, want %q and a remedy", name, err, tc.want)
		}
	}
}
