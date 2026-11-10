package doctor

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestIsMachineCheck(t *testing.T) {
	for _, tc := range []struct {
		check, detail string
		want          bool
	}{
		{"Node.js is 24, as in CI", "found Node.js 22.23.4, older than the 24", true},
		{"the .NET SDK is installed", "dotnet is not on PATH", true},
		{"Go is required", "go is not on PATH", true},
		{"npm workspaces include web", "package.json lacks it", false},
		{"go.work exists", "missing", false},
	} {
		got := IsMachineCheck(Finding{Finding: protocol.Finding{Check: tc.check, Detail: tc.detail}})
		if got != tc.want {
			t.Errorf("%q: got %v, want %v", tc.check, got, tc.want)
		}
	}
}
