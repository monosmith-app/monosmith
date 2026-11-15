package workspace

import (
	"strings"
	"testing"
)

const oldGroupsFile = `schema = 1

[workspace]
name = "acme"

# [[task_groups]] is the old name; this comment must stay
[[task_groups]]
select = ["lang:node"] # libraries
[task_groups.tasks.build]
outputs = ["dist"]
`

func TestRemovedKeysAreFoundInTableHeadersOnly(t *testing.T) {
	if got := FindRemovedKeys([]byte(oldGroupsFile)); len(got) != 1 || got[0].Old != "task_groups" || got[0].New != "groups" {
		t.Fatalf("found %+v", got)
	}
	comment := "# [[task_groups]] was renamed\n[workspace]\nname = \"a\"\n"
	if got := FindRemovedKeys([]byte(comment)); len(got) != 0 {
		t.Fatalf("a comment is not a table: %+v", got)
	}
}

func TestRenameRemovedKeysChangesOnlyTheHeaders(t *testing.T) {
	text, changed := RenameRemovedKeys([]byte(oldGroupsFile))
	if !changed {
		t.Fatal("nothing renamed")
	}
	want := strings.NewReplacer("\n[[task_groups]]", "\n[[groups]]", "[task_groups.tasks.build]", "[groups.tasks.build]").Replace(oldGroupsFile)
	if string(text) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", text, want)
	}
	if !strings.Contains(string(text), "# [[task_groups]] is the old name; this comment must stay") {
		t.Error("the comment changed")
	}
	cfg, err := Parse(text)
	if err != nil || len(cfg.Groups) != 1 || cfg.Groups[0].Tasks["build"].Outputs[0] != "dist" {
		t.Fatalf("the renamed file must load: %+v %v", cfg.Groups, err)
	}
	if _, again := RenameRemovedKeys(text); again {
		t.Error("a renamed file renames nothing")
	}
}

func TestParseSaysOneSentenceForARemovedKey(t *testing.T) {
	_, err := Parse([]byte(oldGroupsFile))
	removed, ok := AsRemovedKeys(err)
	if !ok {
		t.Fatalf("error = %v", err)
	}
	want := "smith.toml uses [[task_groups]], which is now [[groups]]: run smith doctor --fix"
	if removed.Error() != want || strings.Contains(removed.Error(), "\n") {
		t.Fatalf("message = %q", removed.Error())
	}
}

func TestTheSmithPinMovedFromCIToWorkspace(t *testing.T) {
	raw := []byte("[workspace]\nname = \"acme\"\n\n[ci]\nsmith_version = \"2.49.0\"\nproviders = [\"github\"]\n")
	if _, err := Parse(raw); err == nil || !strings.Contains(err.Error(), "[ci] smith_version, which is now [workspace] smith: run smith doctor --fix") {
		t.Fatalf("an old pin must name its new place and the fix, got %v", err)
	}
	fixed, ok := RenameRemovedKeys(raw)
	if !ok {
		t.Fatal("the pin must move")
	}
	cfg, err := Parse(fixed)
	if err != nil || cfg.Workspace.Smith != "2.49.0" || len(cfg.CI.Providers) != 1 {
		t.Fatalf("the moved pin reads as [workspace] smith: %v %+v\n%s", err, cfg.Workspace, fixed)
	}
}
