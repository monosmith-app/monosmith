package splitrepo

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/integrations/gitclient"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := gitclient.New(dir).Run(args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// monorepo is a repository with a PHP package beside another project, tagged
// as smith tags a release; mirror is an empty bare repository.
func monorepo(t *testing.T) (repo, mirror string) {
	t.Helper()
	// git subtree -h prints its usage and exits 129, so read the output, not the status.
	if out, _ := exec.Command("git", "subtree", "-h").CombinedOutput(); !strings.Contains(string(out), "subtree split") {
		t.Skip("git subtree is not installed")
	}
	repo, mirror = t.TempDir(), t.TempDir()
	git(t, mirror, "init", "--bare", "-q")
	git(t, repo, "init", "-q", "-b", "main")
	git(t, repo, "config", "user.email", "smith@example.com")
	git(t, repo, "config", "user.name", "smith")
	git(t, repo, "config", "commit.gpgsign", "false")
	write(t, repo, "packages/client/composer.json", `{"name": "acme/client"}`)
	write(t, repo, "packages/client/src/Client.php", "<?php\n")
	write(t, repo, "apps/web/package.json", "{}")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "feat: client")
	git(t, repo, "tag", "-a", "client@1.2.0", "-m", "client 1.2.0")
	return repo, mirror
}

func TestMirrorPushesTheSplitAndResumes(t *testing.T) {
	repo, mirror := monorepo(t)
	g := gitclient.New(repo)
	req := Request{Prefix: "packages/client", Ref: "client@1.2.0", Repo: "acme/client", URL: mirror, Version: "1.2.0"}

	res, err := Mirror(g, req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Mirror != "acme/client@v1.2.0" || res.Skipped != "" {
		t.Errorf("result %+v", res)
	}
	files := git(t, mirror, "ls-tree", "-r", "--name-only", "v1.2.0")
	if files != "composer.json\nsrc/Client.php" {
		t.Errorf("the mirror holds the package at its root, and nothing else:\n%s", files)
	}
	if git(t, mirror, "rev-parse", "main") != git(t, mirror, "rev-parse", "v1.2.0^{commit}") {
		t.Error("main and the tag are the same split commit")
	}

	again, err := Mirror(g, req)
	if err != nil || again.Skipped != "acme/client already has v1.2.0" {
		t.Errorf("a re-run release leaves the mirror alone: %+v, %v", again, err)
	}

	// The next release fast-forwards the mirror: the same splitter makes the
	// same commits for the history the two releases share.
	write(t, repo, "packages/client/src/Retry.php", "<?php\n")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "-m", "feat: retry")
	git(t, repo, "tag", "-a", "client@1.3.0", "-m", "client 1.3.0")
	if _, err := Mirror(g, Request{Prefix: "packages/client", Ref: "client@1.3.0", Repo: "acme/client", URL: mirror, Version: "1.3.0"}); err != nil {
		t.Fatal(err)
	}
	if git(t, mirror, "rev-parse", "main^") != git(t, mirror, "rev-parse", "v1.2.0^{commit}") {
		t.Error("1.3.0 sits on top of 1.2.0 in the mirror")
	}
}

func TestMirrorSaysWhyAChangedMirrorIsRefused(t *testing.T) {
	repo, mirror := monorepo(t)
	g := gitclient.New(repo)
	// Someone committed to the mirror by hand.
	other := t.TempDir()
	git(t, other, "init", "-q", "-b", "main")
	git(t, other, "config", "user.email", "x@example.com")
	git(t, other, "config", "user.name", "x")
	git(t, other, "config", "commit.gpgsign", "false")
	write(t, other, "README.md", "hand-made\n")
	git(t, other, "add", "-A")
	git(t, other, "commit", "-q", "-m", "by hand")
	git(t, other, "push", "-q", mirror, "main")

	_, err := Mirror(g, Request{Prefix: "packages/client", Ref: "client@1.2.0", Repo: "acme/client", URL: mirror, Version: "1.2.0"})
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Errorf("a mirror changed outside smith: %v", err)
	}
	if out, _ := gitclient.New(mirror).Run("tag"); out != "" {
		t.Errorf("--atomic: the tag is not pushed when main is refused, so it still means done: %q", out)
	}
}

func TestReady(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	if problems := Ready(gitclient.New(t.TempDir())); len(problems) != 0 && !strings.Contains(strings.Join(problems, " "), "subtree") {
		t.Errorf("this machine's git: %v", problems)
	}
	for version, ok := range map[string]bool{
		"git version 2.41.0.windows.1": true,
		"git version 2.31.0":           true,
		"git version 2.30.9":           false,
		"git version 3.0.0":            true,
		"something else":               false,
	} {
		if atLeast(version, 2, 31) != ok {
			t.Errorf("%q: want %v", version, ok)
		}
	}
}

func TestMirrorDryRunAndToken(t *testing.T) {
	res, err := Mirror(gitclient.New(t.TempDir()), Request{Repo: "acme/client", Version: "2.0.0", DryRun: true})
	if err != nil || res.Mirror != "acme/client@v2.0.0" {
		t.Errorf("a dry run only names the mirror: %+v, %v", res, err)
	}
	if _, err := Mirror(gitclient.New(t.TempDir()), Request{Repo: "acme/client", Version: "2.0.0"}); err == nil || !strings.Contains(err.Error(), TokenEnv) {
		t.Errorf("no token: %v", err)
	}
	env := strings.Join(authEnv("s3cret"), "\n")
	if strings.Contains(env, "s3cret") {
		t.Error("the token is encoded in the header, never in clear")
	}
	want := base64.StdEncoding.EncodeToString([]byte("x-access-token:s3cret"))
	if !strings.Contains(env, "GIT_CONFIG_VALUE_0=\n") || !strings.Contains(env, want) {
		t.Errorf("the header list is emptied, then the token's header added:\n%s", env)
	}
}

func TestValid(t *testing.T) {
	for _, ok := range []string{"acme/client", "Open-Craw/opencraw-php", "a.b/c_d"} {
		if err := Valid(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "client", "https://github.com/acme/client", "acme/client/extra", "acme/ client"} {
		if Valid(bad) == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
}
