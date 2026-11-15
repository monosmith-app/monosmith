package guide

import "testing"

func TestTickingOneOptionOfASingleChoiceAxisUnticksTheOther(t *testing.T) {
	opts := []Option{
		{Value: "react", Group: "framework"},
		{Value: "electron", Group: "framework"},
		{Value: "eslint", Group: "lint", Many: true},
		{Value: "biome", Group: "lint", Many: true},
		{Value: "tauri", Group: "shell"},
	}
	marked, note := ToggleOption(nil, opts, 0)
	marked, _ = ToggleOption(marked, opts, 4)
	if note != "" || !marked[0] || !marked[4] {
		t.Fatalf("different axes tick together: %v %q", marked, note)
	}
	marked, note = ToggleOption(marked, opts, 1)
	if marked[0] || !marked[1] || !marked[4] {
		t.Fatalf("electron replaces react and keeps tauri: %v", marked)
	}
	if note == "" {
		t.Fatal("the swap is said in a note")
	}
	marked, _ = ToggleOption(marked, opts, 2)
	marked, note = ToggleOption(marked, opts, 3)
	if !marked[2] || !marked[3] || note != "" {
		t.Fatalf("an additive axis stays checkboxes: %v", marked)
	}
	marked, _ = ToggleOption(marked, opts, 1)
	if marked[1] {
		t.Fatal("ticking a ticked option unticks it")
	}
}
