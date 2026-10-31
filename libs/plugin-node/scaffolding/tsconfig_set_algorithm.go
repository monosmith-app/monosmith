package scaffolding

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
)

// Names of the tsconfig files a project can have.
const (
	// TsconfigBuild is the build configuration smith writes by default.
	TsconfigBuild = "tsconfig.build.json"
	// TsconfigLib is the build configuration of an Nx-style project; tsconfig.json
	// is then a solution file that references it.
	TsconfigLib = "tsconfig.lib.json"
	// TsconfigSpec is the configuration of the project's tests.
	TsconfigSpec = "tsconfig.spec.json"
)

// TsconfigSet reads the tsconfig.files fact of a shape and says which build
// configuration the project has (TsconfigLib when the siblings have one, else
// TsconfigBuild) and whether it has a TsconfigSpec.
func TsconfigSet(shape map[string]string) (build string, spec bool) {
	files := strings.Split(shape[discovery.ShapeTsconfigFiles], ",")
	build = TsconfigBuild
	if slices.Contains(files, TsconfigLib) {
		build = TsconfigLib
	}
	return build, slices.Contains(files, TsconfigSpec)
}
