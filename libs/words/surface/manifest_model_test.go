package surface

import "testing"

func TestAManifestClaimsAnExactKeyAndEveryKeyOfAPrefix(t *testing.T) {
	m := Manifest{Keys: []string{"cmd.add.title"}, Prefixes: []string{"vscode."}}
	for key, want := range map[string]bool{"cmd.add.title": true, "cmd.add.title2": false, "vscode.x.y": true, "jetbrains.x": false, "vscode": false} {
		if got := m.Claims(key); got != want {
			t.Errorf("Claims(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestAnEmptyPrefixClaimsNothing(t *testing.T) {
	if (Manifest{Prefixes: []string{""}}).Claims("any") {
		t.Fatal("an empty prefix must not claim every key")
	}
}
