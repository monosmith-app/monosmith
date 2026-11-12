package eslintpeer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMeets(t *testing.T) {
	for version, want := range map[string]bool{
		"9.39.3": false, "10.0.0": false, "10.8.0": false, "10.8.1": true, "10.12.0": true, "11.0.0": false,
	} {
		if got := Meets(version); got != want {
			t.Errorf("Meets(%q) = %v, want %v", version, got, want)
		}
	}
}

func TestRaised(t *testing.T) {
	for _, tc := range []struct {
		current, want string
		raised        bool
	}{
		{"^9.39.3", Range, true},
		{"~8.57.0", Range, true},
		{"9", Range, true},
		{"^11.0.0", Range, true},
		{"^10.8.1", "^10.8.1", false},
		{"^10.12.0", "^10.12.0", false},
		{"latest", "latest", false},
		{"$eslint", "$eslint", false},
		{"^9 || ^10", "^9 || ^10", false},
	} {
		got, raised := Raised(tc.current)
		if got != tc.want || raised != tc.raised {
			t.Errorf("Raised(%q) = %q, %v; want %q, %v", tc.current, got, raised, tc.want, tc.raised)
		}
	}
}

// The range is the config package's own peer dependency: when the package
// moves to a newer ESLint, this fails until the constants follow.
func TestRangeIsTheConfigPackagesPeer(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "eslint-config", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name  string            `json:"name"`
		Peers map[string]string `json:"peerDependencies"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != ConfigPackage || manifest.Peers["eslint"] != Range || "^"+Floor != Range {
		t.Fatalf("%s peers eslint %q; Range = %q, Floor %q", manifest.Name, manifest.Peers["eslint"], Range, Floor)
	}
}
