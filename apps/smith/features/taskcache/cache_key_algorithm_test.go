package taskcache

import "testing"

func TestTheToolchainIsPartOfTheKeyOnlyWhenSet(t *testing.T) {
	base := KeyInputs{Task: "web:build", Command: []string{"npm", "run", "build"}, Platform: "linux/amd64", Tool: "1.0.0/node"}
	plain := Key(base)
	node20, node22 := base, base
	node20.Toolchain, node22.Toolchain = "node 20.11.0", "node 22.3.0"
	if Key(node20) == plain || Key(node20) == Key(node22) {
		t.Error("a reported toolchain changes the key, and another version changes it again")
	}
	again := node20
	if Key(node20) != Key(again) {
		t.Error("the key is stable")
	}
	base.Toolchain = ""
	if Key(base) != plain {
		t.Error("an unreported toolchain leaves the key as it was")
	}
}
