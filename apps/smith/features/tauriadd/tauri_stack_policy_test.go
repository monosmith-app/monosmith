package tauriadd

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func lookupIn(stacks ...protocol.Stack) func(string) (protocol.Stack, bool) {
	return func(name string) (protocol.Stack, bool) {
		for _, s := range stacks {
			if s.Name == name {
				return s, true
			}
		}
		return protocol.Stack{}, false
	}
}

func TestDeriveStackReusesAStackWithTheSameMakeUp(t *testing.T) {
	react := protocol.Stack{Name: "react", Language: "node", Scaffold: "vite", Toolchain: []string{"vite", "eslint", "vitest"}}
	reactTauri := protocol.Stack{Name: "react-tauri", Language: "node", Scaffold: "vite", Toolchain: []string{"vite", "eslint", "vitest", "tauri"}}
	st, declare := DeriveStack(react, "studio", lookupIn(react, reactTauri))
	if st.Name != "react-tauri" || declare {
		t.Errorf("got %s declare=%v", st.Name, declare)
	}
}

func TestDeriveStackKeepsTheProjectsOwnTooling(t *testing.T) {
	jest := protocol.Stack{Name: "node-jest", Language: "node", Toolchain: []string{"typescript", "eslint", "jest"}}
	st, declare := DeriveStack(jest, "studio", lookupIn(jest))
	if st.Name != "node-jest-tauri" || !declare || !slices.Equal(st.Toolchain, []string{"typescript", "eslint", "jest", "tauri"}) {
		t.Errorf("got %+v declare=%v", st, declare)
	}
}

func TestDeriveStackNamesItAfterTheProjectWhenTheNameIsTaken(t *testing.T) {
	jest := protocol.Stack{Name: "node-jest", Language: "node", Toolchain: []string{"eslint", "jest"}}
	other := protocol.Stack{Name: "node-jest-tauri", Language: "node", Toolchain: []string{"jest"}}
	st, declare := DeriveStack(jest, "studio", lookupIn(jest, other))
	if st.Name != "studio-tauri" || !declare {
		t.Errorf("got %s declare=%v", st.Name, declare)
	}
}

func TestDeriveStackFromNoStackAndWhenTauriIsThere(t *testing.T) {
	st, declare := DeriveStack(protocol.Stack{}, "studio", lookupIn())
	if st.Name != "node-tauri" || !declare || !slices.Equal(st.Toolchain, []string{"tauri"}) {
		t.Errorf("got %+v", st)
	}
	has := protocol.Stack{Name: "x", Language: "node", Toolchain: []string{"tauri"}}
	if st, declare := DeriveStack(has, "p", lookupIn(has)); st.Name != "x" || declare {
		t.Errorf("got %+v", st)
	}
}
