package compose

import "testing"

func TestRivalsAreTheTickedChoicesOnTheSameSingleChoiceAxis(t *testing.T) {
	react := Item{ID: "node/react", Kind: KindChoice, Name: "react", Axis: "framework", Language: "node"}
	electron := Item{ID: "node/electron", Kind: KindChoice, Name: "electron", Axis: "framework", Language: "node"}
	tauri := Item{ID: "node/tauri", Kind: KindChoice, Name: "tauri", Axis: "shell", Language: "node"}
	vitest := Item{ID: "node/vitest", Kind: KindChoice, Name: "vitest", Axis: "test", Language: "node", Many: true}
	jest := Item{ID: "node/jest", Kind: KindChoice, Name: "jest", Axis: "test", Language: "node", Many: true}
	cases := []struct {
		name   string
		ticked []Item
		pick   Item
		want   []string
	}{
		{"same axis", []Item{react, tauri}, electron, []string{"react"}},
		{"other axis", []Item{react}, tauri, nil},
		{"both many share", []Item{vitest}, jest, nil},
		{"one many one single does not", []Item{vitest}, Item{ID: "node/x", Kind: KindChoice, Name: "x", Axis: "test", Language: "node"}, []string{"vitest"}},
		{"the same item", []Item{react}, react, nil},
		{"a language has none", []Item{react}, Item{ID: "node", Kind: KindLanguage, Name: "node"}, nil},
	}
	for _, c := range cases {
		var got []string
		for _, r := range Rivals(c.ticked, c.pick) {
			got = append(got, r.Name)
		}
		if len(got) != len(c.want) || (len(got) > 0 && got[0] != c.want[0]) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
