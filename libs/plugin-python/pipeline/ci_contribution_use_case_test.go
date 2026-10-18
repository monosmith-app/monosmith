package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIPointsTheVenvsPipAtThePrivateIndexBeforeInstalling(t *testing.T) {
	feed := &protocol.Registry{Name: "libs", Auth: protocol.AuthAzureArtifacts, PyPI: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/", PyPIUpload: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/upload/"}
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: feed, Publish: feed}}}})
	if len(r.Setup) != 4 || r.Setup[1].Name != "Point pip at the private index" {
		t.Fatalf("setup: %+v", r.Setup)
	}
	if got := strings.Join(r.Setup[1].Command[1:], " "); got != "-m pip config --site set global.index-url "+feed.PyPI {
		t.Errorf("command %q", got)
	}
	for _, c := range r.Credentials {
		if c.Env == "PYPI_TOKEN" {
			t.Error("published to Azure Artifacts: no PyPI token")
		}
	}
	if len(r.Authenticate) != 2 {
		t.Errorf("sign-ins: %+v", r.Authenticate)
	}
}

func TestCIWithoutARegistryIsUnchanged(t *testing.T) {
	r := CI(protocol.CIParams{Projects: []protocol.Project{{Name: "sdk", Root: "packages/sdk", Kind: protocol.ProjectPackage}}})
	if len(r.Setup) != 3 || len(r.Authenticate) != 0 || len(r.Credentials) != 1 || r.Credentials[0].Env != "PYPI_TOKEN" {
		t.Errorf("result %+v", r)
	}
}
