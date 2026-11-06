package discovery

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestWithDeclaredAppendsInOrder(t *testing.T) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{{Name: "found"}}}
	declared := []protocol.DeclaredProject{{Name: "a", Root: "x/a"}, {Name: "b", Root: "x/b"}}
	got, err := WithDeclared(result, declared, func(d protocol.DeclaredProject) (protocol.Project, error) {
		return protocol.Project{Name: d.Name}, nil
	})
	if err != nil || len(got.Projects) != 3 || got.Projects[1].Name != "a" || got.Projects[2].Name != "b" {
		t.Fatalf("got %+v, err %v", got.Projects, err)
	}
}

func TestWithDeclaredNamesTheFailedDeclaration(t *testing.T) {
	cause := errors.New("no manifest")
	got, err := WithDeclared(protocol.DiscoverResult{}, []protocol.DeclaredProject{{Name: "a", Root: "x/a"}}, func(protocol.DeclaredProject) (protocol.Project, error) {
		return protocol.Project{}, cause
	})
	if !errors.Is(err, cause) || err.Error() != "[projects.a] at x/a: no manifest" || len(got.Projects) != 0 {
		t.Fatalf("got %+v, err %v", got.Projects, err)
	}
}
