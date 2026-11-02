package linking

import "testing"

func TestAddRequirePutsTheModuleInTheRequireBlock(t *testing.T) {
	src := "module example.com/r/apps/api\n\ngo 1.26\n\nrequire (\n\tgithub.com/x/y v1.2.3\n)\n"
	got, changed := AddRequire(src, "example.com/r/libs/core", "")
	want := "module example.com/r/apps/api\n\ngo 1.26\n\nrequire (\n\tgithub.com/x/y v1.2.3\n\texample.com/r/libs/core v0.0.0\n)\n"
	if !changed || got != want {
		t.Fatalf("got:\n%s", got)
	}
	if again, changed := AddRequire(got, "example.com/r/libs/core", ""); changed || again != got {
		t.Fatal("a second add changes nothing")
	}
}

func TestAddRequireAppendsALineAndKeepsCRLF(t *testing.T) {
	got, changed := AddRequire("module m\r\n\r\ngo 1.26\r\n", "example.com/core", "v1.0.0")
	if !changed || got != "module m\r\n\r\ngo 1.26\r\n\r\nrequire example.com/core v1.0.0\r\n" {
		t.Fatalf("got %q", got)
	}
}

func TestRemoveRequireTakesTheReplaceToo(t *testing.T) {
	src := "module m\n\nrequire example.com/core v0.0.0\n\nrequire (\n\tgithub.com/x/y v1.2.3\n)\n\nreplace example.com/core => ../../libs/core\n"
	got, changed := RemoveRequire(src, "example.com/core")
	want := "module m\n\n\nrequire (\n\tgithub.com/x/y v1.2.3\n)\n\n"
	if !changed || got != want {
		t.Fatalf("got %q", got)
	}
	if _, changed := RemoveRequire(got, "example.com/core"); changed {
		t.Fatal("nothing left to remove")
	}
}

func TestAddReplaceOnce(t *testing.T) {
	got, changed := AddReplace("module m\n", "example.com/core", "../../libs/core")
	if !changed || got != "module m\n\nreplace example.com/core => ../../libs/core\n" {
		t.Fatalf("got %q", got)
	}
	if _, changed := AddReplace(got, "example.com/core", "../x"); changed {
		t.Fatal("already replaced")
	}
}

func TestWorkUsesReadsBlocksAndSingleLines(t *testing.T) {
	work := "go 1.26\n\nuse (\n\t./apps/api // the api\n\t\"./libs/core\"\n)\n\nuse ./libs/ui\n"
	for folder, want := range map[string]bool{"apps/api": true, "libs/core": true, "libs/ui": true, "libs/other": false} {
		if got := WorkUses(work, folder); got != want {
			t.Errorf("WorkUses(%q) = %v", folder, got)
		}
	}
}

func TestRelativeFolder(t *testing.T) {
	if got := relativeFolder("apps/api", "libs/core"); got != "../../libs/core" {
		t.Errorf("got %q", got)
	}
	if got := relativeFolder("libs", "libs/core"); got != "./core" {
		t.Errorf("got %q", got)
	}
}
