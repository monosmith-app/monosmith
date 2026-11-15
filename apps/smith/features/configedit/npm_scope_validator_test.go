package configedit

import "testing"

func TestTheWorkspaceScopeMustBeAnNpmScope(t *testing.T) {
	f := Field{Path: "workspace.scope", Type: "string"}
	for scope, ok := range map[string]bool{
		"@acme": true, "@my-org": true, "@a.b_c": true, "": true,
		"Bad Scope!": false, "acme": false, "@Acme": false, "@": false, "@acme/x": false,
	} {
		_, err := f.Encode([]string{scope})
		if (err == nil) != ok {
			t.Errorf("scope %q: err = %v, want accepted = %v", scope, err, ok)
		}
	}
}
