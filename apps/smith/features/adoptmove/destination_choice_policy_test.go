package adoptmove

import "testing"

func TestChoicesSuggestAnAppFolderForAProgram(t *testing.T) {
	for _, c := range []struct {
		app  bool
		want int
	}{{true, 1}, {false, 2}} {
		choices := Choices(Candidate{Name: "web", App: c.app})
		if len(choices) != 5 {
			t.Fatalf("got %d choices, want 5", len(choices))
		}
		if got := DefaultChoice(choices); got != c.want {
			t.Errorf("app=%v: default %d, want %d", c.app, got, c.want)
		}
		if choices[0].Dir != "apps/web" || choices[1].Dir != "packages/web" || choices[2].Dir != "libs/web" {
			t.Errorf("folders = %v", choices)
		}
		if choices[3].Key != KeyOther || choices[4].Key != KeyNone {
			t.Errorf("last two = %v", choices[3:])
		}
	}
}

func TestDestination(t *testing.T) {
	cases := []struct {
		key, name, want string
		fails           bool
	}{
		{KeyApps, "web", "apps/web", false},
		{KeyLibs, "core-kit", "libs/core-kit", false},
		{KeyPackages, "Web", "", true},
		{KeyApps, "web_app", "", true},
		{KeyOther, "tools/web", "tools/web", false},
		{KeyOther, ".", "", true},
		{KeyOther, "../web", "", true},
		{KeyOther, "/abs/web", "", true},
		{KeyOther, `C:\web`, "", true},
		{KeyOther, ".github/web", "", true},
		{KeyOther, "a b", "", true},
		{KeyOther, `tools\web`, "tools/web", false},
		{KeyNone, "web", "", true},
	}
	for _, c := range cases {
		got, err := Destination(c.key, c.name)
		if (err != nil) != c.fails || got != c.want {
			t.Errorf("Destination(%q, %q) = %q, %v; want %q (fails=%v)", c.key, c.name, got, err, c.want, c.fails)
		}
	}
}
