package commands

import (
	"strings"
	"testing"
)

var (
	inside  = State{InWorkspace: true, HasSelection: true}
	outside = State{}
	busy    = State{InWorkspace: true, Running: true, HasSelection: true}
)

func TestNoKeyIsBoundTwiceInAnyState(t *testing.T) {
	for _, s := range []State{inside, outside, busy} {
		seen := map[string]string{}
		for _, c := range all {
			if !c.Visible(s) {
				continue
			}
			for _, k := range c.Keys {
				if prev, dup := seen[k]; dup {
					t.Errorf("%q is bound to %s and %s", k, prev, c.ID)
				}
				seen[k] = c.ID
			}
		}
	}
}

func TestEveryCommandWithAKeyShowsItsShortcut(t *testing.T) {
	for _, c := range all {
		if len(c.Keys) > 0 && c.Shortcut == "" {
			t.Errorf("%s has keys but shows no shortcut", c.ID)
		}
		if c.Shortcut != "" && !strings.EqualFold(c.Shortcut, c.Keys[0]) {
			t.Errorf("%s shows %q but its first key is %q", c.ID, c.Shortcut, c.Keys[0])
		}
	}
}

func TestEveryMenuCommandHasATitleAndExecuteHasAnArgv(t *testing.T) {
	for _, c := range all {
		if c.Title == "" || c.ID == "" {
			t.Errorf("%+v", c)
		}
		if c.Kind == Execute && len(c.Argv) == 0 {
			t.Errorf("%s executes nothing", c.ID)
		}
		if c.Menu != "" {
			if _, ok := Mnemonics[c.Menu]; !ok {
				t.Errorf("%s is in an unknown menu %q", c.ID, c.Menu)
			}
		}
	}
}

func TestOutsideAWorkspaceTheWorkspaceCommandsSayWhyTheyAreOff(t *testing.T) {
	c, reason, ok := ByKey("ctrl+d", outside)
	if !ok || c.ID != "ws.doctor" || !strings.Contains(reason, "not in a workspace") {
		t.Fatalf("%v %q %v", c.ID, reason, ok)
	}
	if _, reason, _ := ByKey("ctrl+d", inside); reason != "" {
		t.Fatalf("available inside: %q", reason)
	}
}

func TestWhileARunIsGoingNothingElseStartsAndStopIsAvailable(t *testing.T) {
	if _, reason, _ := ByKey("ctrl+b", busy); !strings.Contains(reason, "run is in progress") {
		t.Fatalf("%q", reason)
	}
	if _, reason, _ := ByKey("ctrl+c", busy); reason != "" {
		t.Fatalf("stop should be available: %q", reason)
	}
	if _, reason, _ := ByKey("ctrl+c", inside); reason != "nothing is running" {
		t.Fatalf("%q", reason)
	}
}

func TestRunSelectionNeedsASelection(t *testing.T) {
	if _, reason, _ := ByKey("ctrl+r", State{InWorkspace: true}); !strings.Contains(reason, "select a task in the navigator") {
		t.Fatalf("%q", reason)
	}
}

func TestNewWorkspaceIsOnlyOutsideAndAddProjectOnlyInside(t *testing.T) {
	if c, _, _ := ByKey("ctrl+n", outside); c.ID != "file.new" {
		t.Fatalf("outside: %s", c.ID)
	}
	if c, _, _ := ByKey("ctrl+n", inside); c.ID != "file.add" {
		t.Fatalf("inside: %s", c.ID)
	}
}

func TestTheMenusListEveryShortcutAndDisableWithAReason(t *testing.T) {
	menus := Menus(outside)
	var names []string
	for _, m := range menus {
		names = append(names, m.Name)
	}
	if got := strings.Join(names, ","); got != "File,Workspace,Run,Projects,Release,View,Help" {
		t.Fatalf("menus %s", got)
	}
	for _, m := range menus {
		for _, e := range m.Entries {
			if e.Separator {
				continue
			}
			if e.ID == "ws.doctor" && !strings.Contains(e.Disabled, "not in a workspace") {
				t.Errorf("doctor outside should say why: %+v", e)
			}
			if e.ID == "file.quit" && (e.Shortcut != "Ctrl+C" || e.Disabled != "") {
				t.Errorf("quit: %+v", e)
			}
		}
	}
}

func TestShortcutsListsOnlyVisibleCommands(t *testing.T) {
	for _, c := range Shortcuts(inside) {
		if c.ID == "file.new" {
			t.Fatal("New workspace is not offered inside a workspace")
		}
	}
}
