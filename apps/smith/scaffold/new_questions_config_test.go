package scaffold

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
)

func TestEveryArchitectureHasAShortHintAndTheDefaultExists(t *testing.T) {
	for _, a := range architecture.Catalog() {
		if hint := layoutHints[a.Name]; hint == "" || len(hint) > 50 {
			t.Errorf("%s: hint %q is missing or too long for a line", a.Name, hint)
		}
	}
	if _, err := architecture.Find(DefaultArchitecture); err != nil {
		t.Error(err)
	}
}
