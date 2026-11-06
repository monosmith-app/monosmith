package splitter

import "testing"

func TestIntentOfReadsTheDividerKeys(t *testing.T) {
	for name, want := range map[string]KeyIntent{
		`\`:           {Action: CycleKey, Dir: 1},
		"|":           {Action: CycleKey, Dir: -1},
		"shift+left":  {Action: SelectKey, Axis: Column, Dir: -1},
		"shift+right": {Action: SelectKey, Axis: Column, Dir: 1},
		"shift+up":    {Action: SelectKey, Axis: Row, Dir: -1},
		"shift+down":  {Action: SelectKey, Axis: Row, Dir: 1},
		"tab":         {Action: AdvanceKey, Dir: 1},
		"esc":         {Action: ReleaseKey},
		"x":           {},
	} {
		if got := IntentOf(name); got != want {
			t.Errorf("IntentOf(%q) = %+v, want %+v", name, got, want)
		}
	}
}
