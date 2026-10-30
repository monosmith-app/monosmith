package declaredimports

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDoctorReportsEachModule(t *testing.T) {
	root := plainModule(t)
	write(t, root, "bad/go.mod", "module example.com/bad\n\ngo 1.22\n\nrequire example.com/gone v1.0.0\n")
	write(t, root, "bad/b.go", "package bad\n")
	result := Doctor(root, protocol.DoctorParams{Projects: []protocol.Project{{Name: "m", Root: "m"}, {Name: "bad", Root: "bad"}}})
	if len(result.Findings) != 2 {
		t.Fatalf("got %+v", result.Findings)
	}
	if !result.Findings[0].OK {
		t.Errorf("m is tidy: %+v", result.Findings[0])
	}
	bad := result.Findings[1]
	if bad.OK || !strings.Contains(bad.Detail, "example.com/gone") || !strings.Contains(bad.Remedy, "go mod tidy") {
		t.Errorf("got %+v", bad)
	}
}
