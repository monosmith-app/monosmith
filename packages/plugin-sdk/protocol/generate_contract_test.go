package protocol

import "testing"

func TestPackagePrefixIsTheScopeWithoutItsAtElseTheName(t *testing.T) {
	for _, c := range []struct {
		ws   WorkspaceInfo
		want string
	}{
		{WorkspaceInfo{Name: "opencraw", Scope: "@pargic"}, "pargic"},
		{WorkspaceInfo{Name: "opencraw", Scope: "pargic"}, "pargic"},
		{WorkspaceInfo{Name: "opencraw", Scope: " @ "}, "opencraw"},
		{WorkspaceInfo{Name: "opencraw"}, "opencraw"},
		{WorkspaceInfo{}, ""},
	} {
		if got := c.ws.PackagePrefix(); got != c.want {
			t.Errorf("%+v: got %q, want %q", c.ws, got, c.want)
		}
	}
}
