package words

import (
	"testing"

	"github.com/monosmith-app/monosmith/libs/words/surface"
)

// A surface that lists a key the catalogue lacks would ship the key name on screen.
func TestEveryExactKeyASurfaceClaimsIsInTheCatalogue(t *testing.T) {
	en := Reference()
	for _, name := range surface.Names {
		m, err := surface.Load(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range m.Keys {
			if _, ok := en[key]; !ok {
				t.Errorf("surface %s claims %q, which the catalogue lacks", name, key)
			}
		}
	}
}
