package registries

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAPackageThatPublishesToNoRegistryAsksNoPublicCredential(t *testing.T) {
	private := &protocol.Registry{Name: "feed", Auth: protocol.AuthToken, TokenEnv: "FEED_TOKEN", NPM: "https://feed.example/npm/"}
	for name, mode := range map[string]*protocol.Registries{
		"none":           {NoPublish: true},
		"github-release": {PackToRelease: true},
	} {
		projects := []protocol.Project{
			{Name: "a", Kind: protocol.ProjectPackage, Registries: mode},
			{Name: "b", Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Publish: private}},
		}
		if _, public := CI(projects); public {
			t.Errorf("%s: a project that publishes nowhere must not ask for the public npm token", name)
		}
	}
	if _, public := CI([]protocol.Project{{Name: "c", Kind: protocol.ProjectPackage}}); !public {
		t.Error("a package with no setting still publishes to the public registry")
	}
}
