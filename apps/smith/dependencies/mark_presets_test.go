package dependencies

import "testing"

func TestPresetsComeFirstAsOneGroup(t *testing.T) {
	updates := []Update{
		{Name: "zod", Level: "patch"},
		{Name: "@acme/eslint-config", Level: "major"},
		{Name: "jest", Level: "minor"},
	}
	MarkPresets(updates, []string{"@acme/eslint-config"})
	if updates[0].Name != "@acme/eslint-config" || !updates[0].Preset || updates[0].Group() != "preset" {
		t.Fatalf("got %+v", updates)
	}
	if updates[1].Name != "zod" || updates[1].Group() != "patch" || updates[2].Name != "jest" {
		t.Errorf("the others changed order or group: %+v", updates)
	}
}
