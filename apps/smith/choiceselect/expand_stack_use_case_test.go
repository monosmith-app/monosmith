package choiceselect

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestExpandStackDerivesTheStackFromTheSelection(t *testing.T) {
	st, err := ExpandStack(protocol.Stack{Name: "web", Choices: []string{"react", "bootstrap"}, Toolchain: []string{"extra"}, Deploy: []string{"d"}}, webOffers())
	if err != nil {
		t.Fatal(err)
	}
	if st.Language != "node" || st.Scaffold != "vite" || !slices.Equal(st.Toolchain, []string{"vite", "eslint", "vitest", "extra"}) || !slices.Equal(st.Deploy, []string{"d"}) {
		t.Errorf("stack %+v", st)
	}
	if !slices.Equal(st.Choices, []string{"react", "bootstrap", "eslint", "vitest"}) {
		t.Errorf("the defaults are part of the saved choices: %v", st.Choices)
	}
	if _, err := ExpandStack(protocol.Stack{Name: "web", Choices: []string{"react"}, Scaffold: "other"}, webOffers()); err == nil || !strings.Contains(err.Error(), "scaffold other") {
		t.Errorf("a scaffold the choices contradict: %v", err)
	}
	if _, err := ExpandStack(protocol.Stack{Name: "web", Choices: []string{"react"}, Language: "go"}, webOffers()); err == nil || !strings.Contains(err.Error(), "language go") {
		t.Errorf("a language the choices contradict: %v", err)
	}
	if plain, err := ExpandStack(protocol.Stack{Name: "p", Language: "node"}, webOffers()); err != nil || plain.Language != "node" {
		t.Errorf("a stack with no choices is unchanged: %+v %v", plain, err)
	}
}
