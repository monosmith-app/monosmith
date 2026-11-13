package typescript

import (
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// A change of the workspace's tsconfig.base.json must reach every project the
// stack writes, for every kind and for the Nx-style set of tsconfig files.
func TestEveryTsconfigOfTheStackReachesTheBase(t *testing.T) {
	for name, c := range map[string]struct {
		kind  protocol.ProjectKind
		shape map[string]string
	}{
		"app":     {protocol.ProjectApp, nil},
		"lib":     {protocol.ProjectLib, nil},
		"package": {protocol.ProjectPackage, nil},
		"nx set":  {protocol.ProjectLib, map[string]string{discovery.ShapeTsconfigFiles: "tsconfig.json,tsconfig.lib.json,tsconfig.spec.json"}},
	} {
		t.Run(name, func(t *testing.T) {
			ws := generatedFiles(t, c.kind, c.shape)
			if read(t, ws, "packages/dep/tsconfig.json") == "" {
				t.Fatal("the stack writes a tsconfig.json")
			}
			if read(t, ws, tsconfigchain.Base) != tsconfigchain.BaseFile {
				t.Error("the base extends the preset")
			}
			if bad, err := tsconfigchain.Unreached(ws); err != nil || len(bad) > 0 {
				t.Errorf("these tsconfig files do not reach tsconfig.base.json: %v (%v)", bad, err)
			}
		})
	}
}
