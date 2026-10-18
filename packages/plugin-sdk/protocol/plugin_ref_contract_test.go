package protocol

import "testing"

func TestPluginRefNamesWhatItIsAbout(t *testing.T) {
	for _, c := range []struct {
		ref  PluginRef
		want string
	}{
		{RequiresLanguage("node"), "language plugin node"},
		{ConflictsWith("jest"), "plugin jest"},
		{PluginRef{Level: RoleToolchain, Slot: "test"}, "toolchain plugin providing test"},
	} {
		if got := c.ref.String(); got != c.want {
			t.Errorf("%+v: %q, want %q", c.ref, got, c.want)
		}
	}
	if !(PluginRef{}).IsZero() || RequiresLanguage("node").IsZero() {
		t.Error("IsZero")
	}
}

func TestRolesAreExclusiveOrAdditive(t *testing.T) {
	for _, r := range Roles {
		if r.Exclusive() && r.Additive() {
			t.Errorf("%s is both exclusive and additive", r)
		}
	}
	if !RoleLanguage.Exclusive() || !RoleScaffold.Exclusive() || !RoleStructure.Additive() || !RoleDeploy.Additive() {
		t.Error("levels")
	}
	if RoleToolchain.Exclusive() || RoleToolchain.Additive() {
		t.Error("a toolchain is exclusive per slot, neither wholly")
	}
	if Role("platform").Known() || !RoleDeploy.Known() {
		t.Error("Known")
	}
}
