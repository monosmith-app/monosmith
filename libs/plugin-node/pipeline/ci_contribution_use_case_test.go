package pipeline

import (
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIDeclaresTheNpmDownloadStore(t *testing.T) {
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}}, "22").Stores
	want := []protocol.PackageStore{{Name: "npm", Path: "~/.npm", Env: "NPM_CONFIG_CACHE", Files: []string{"package-lock.json"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%+v", got)
	}
	if none := CI(t.TempDir(), protocol.CIParams{}, "22"); len(none.Stores) != 0 {
		t.Errorf("no Node project, no store: %+v", none.Stores)
	}
}
