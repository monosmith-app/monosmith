package guide

import "testing"

func TestOneEntryBringsARepositoryIn(t *testing.T) {
	var wizards []string
	for _, a := range Actions([]string{"go", "node"}, nil, nil) {
		if a.Wizard == WizardAdopt {
			wizards = append(wizards, a.Title)
		}
	}
	if len(wizards) != 1 || wizards[0] != "Bring this repository into MonoSmith" {
		t.Fatalf("the wizard has one entry, got %q", wizards)
	}
}
