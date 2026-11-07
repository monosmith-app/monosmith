package words

import (
	"strings"
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

// A surface that claims one form of a plural by name must claim the other too,
// or the code that picks the form would find a key the surface never received.
func TestASurfaceThatClaimsAPluralFormClaimsItsPair(t *testing.T) {
	for _, name := range surface.Names {
		m, err := surface.Load(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range m.Keys {
			var pair string
			switch {
			case strings.HasSuffix(key, ".one"):
				pair = strings.TrimSuffix(key, ".one") + ".other"
			case strings.HasSuffix(key, ".other"):
				pair = strings.TrimSuffix(key, ".other") + ".one"
			default:
				continue
			}
			if !m.Claims(pair) {
				t.Errorf("surface %s claims %q but not %q", name, key, pair)
			}
		}
	}
}
