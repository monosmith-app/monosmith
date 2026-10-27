package pathrefs

import "testing"

func TestRetargetRewritesEveryWayAFileNamesTheDirectory(t *testing.T) {
	cases := []struct {
		name, text, fileDir, want string
		count                     int
	}{
		{"workspace file, plain", `members = ["libs/core", "libs/other"]`, "", `members = ["packages/core", "libs/other"]`, 1},
		{"workspace file, dot slash", "use ./libs/core\n", "", "use ./packages/core\n", 1},
		{"sibling", `core = { path = "../core" }`, "libs/b", `core = { path = "../../packages/core" }`, 1},
		{"two levels up", `core = { path = "../../libs/core" }`, "apps/web", `core = { path = "../../packages/core" }`, 1},
		{"subpath", `"paths": {"@x/core": ["libs/core/src"]}`, "", `"paths": {"@x/core": ["packages/core/src"]}`, 1},
		{"longer name is another project", `a = "libs/core-extra" b = "libs/core.old" c = "mylibs/core"`, "", `a = "libs/core-extra" b = "libs/core.old" c = "mylibs/core"`, 0},
		{"a path that leads up to it is another place", `a = "../libs/core"`, "", `a = "../libs/core"`, 0},
		{"twice", "libs/core libs/core\n", "", "packages/core packages/core\n", 2},
	}
	for _, c := range cases {
		got, n := Retarget(c.text, c.fileDir, "libs/core", "packages/core")
		if got != c.want || n != c.count {
			t.Errorf("%s: got %q (%d), want %q (%d)", c.name, got, n, c.want, c.count)
		}
	}
}

func TestRetargetSiblingStaysWhenOnlyTheFileMoves(t *testing.T) {
	// A reference from a file inside libs/a to its sibling libs/b, once libs/a
	// itself moves, is Rebase's job, not Retarget's.
	if got := Rebase("../b", "libs/a", "packages/a"); got != "../../libs/b" {
		t.Fatalf("rebased sibling: %q", got)
	}
	if got := Rebase("../../tsconfig.base.json", "libs/a", "packages/a"); got != "../../tsconfig.base.json" {
		t.Fatalf("rebased shared file: %q", got)
	}
	if got := Rebase("./vendor/x", "libs/a", "packages/a"); got != "./vendor/x" {
		t.Fatalf("a path inside the project moves with it: %q", got)
	}
	if got := Rebase("serde", "libs/a", "packages/a"); got != "serde" {
		t.Fatalf("not a path: %q", got)
	}
}
