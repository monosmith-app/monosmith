package surface

import "testing"

func TestEveryNamedSurfaceHasAnEmbeddedManifest(t *testing.T) {
	for _, name := range Names {
		if _, err := Load(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestTheEditorSurfacesClaimTheirOwnPrefix(t *testing.T) {
	for name, key := range map[string]string{"vscode": "vscode.x", "jetbrains": "jetbrains.x"} {
		m, err := Load(name)
		if err != nil || !m.Claims(key) {
			t.Errorf("%s must claim %q (err %v)", name, key, err)
		}
	}
}

func TestParseRejectsAnUnknownField(t *testing.T) {
	if _, err := Parse([]byte(`{"keys": [], "prefix": []}`)); err == nil {
		t.Fatal("a misspelt field must be an error")
	}
}
