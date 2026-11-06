package toolinstall

import (
	"slices"
	"testing"
)

func TestCargoInstallsAProgramWithTheLockfile(t *testing.T) {
	cmd, ok := BuildCommand("cargo", "cargo-audit", Env{OS: "linux"})
	if want := []string{"cargo", "install", "--locked", "cargo-audit"}; !ok || !slices.Equal(cmd.Argv, want) || cmd.Admin {
		t.Fatalf("got %v admin=%v ok=%v, want %v and no administrator rights", cmd.Argv, cmd.Admin, ok, want)
	}
}

func TestCargoIsOfferedOnlyWhenItIsOnTheMachine(t *testing.T) {
	offered := func(look LookPath) bool {
		for _, m := range DetectManagers("linux", look) {
			if m.Name == "cargo" {
				return true
			}
		}
		return false
	}
	if !offered(onPath("cargo")) {
		t.Error("cargo is on the machine, but it is not offered as a way to install a cargo program")
	}
	if offered(onPath()) {
		t.Error("cargo is offered although it is not on the machine")
	}
}
