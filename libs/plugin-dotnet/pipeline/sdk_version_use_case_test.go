package pipeline

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorComparesTheSDKWithCI(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	ws := t.TempDir()
	params := protocol.DoctorParams{Projects: []protocol.Project{{Root: "apps/api"}}}
	find := func() *protocol.Finding {
		for _, f := range Doctor(ws, params).Findings {
			if f.Check == ".NET SDK is 10.0, as in CI" {
				return &f
			}
		}
		return nil
	}
	for _, c := range []struct {
		out      string
		ok, warn bool
	}{{"10.0.100\n", true, false}, {"8.0.404\n", false, false}, {"11.0.100\n", false, true}} {
		probe = func(string, []string) (string, error) { return c.out, nil }
		if got := find(); got == nil || got.OK != c.ok || got.Warning != c.warn {
			t.Errorf("%q: %+v", c.out, got)
		}
	}
	probe = func(string, []string) (string, error) { return "", errors.New("missing") }
	if find() != nil {
		t.Fatal("a missing dotnet is the install check's finding")
	}
}
