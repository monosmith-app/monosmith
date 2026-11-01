package nativebuild

import (
	"runtime"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAnAppWithItsOwnBuildCommandHasNothingForANativeJobToAttach(t *testing.T) {
	app := func(r *workspace.ProjectRelease) *projectgraph.Project {
		return &projectgraph.Project{Project: protocol.Project{Name: "app", Root: "apps/app", Kind: protocol.ProjectApp}, Release: r}
	}
	if spec := platformAssetSpec(app(&workspace.ProjectRelease{Build: "make release", Assets: []string{"out/*"}})); spec != nil {
		t.Fatalf("a build command makes its own files, the assets job attaches them: %+v", spec)
	}
	spec := platformAssetSpec(app(&workspace.ProjectRelease{Product: "app"}))
	if spec == nil || len(spec.Platforms) != 1 || spec.Platforms[0] != runtime.GOOS+"/"+runtime.GOARCH {
		t.Fatalf("a plugin-built app is built for this platform only: %+v", spec)
	}
}
