package releasing

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(r *protocol.Registry) protocol.Project {
	return protocol.Project{Name: "sdk", Registries: &protocol.Registries{Publish: r}}
}

func TestAzureArtifactsUploadsWithTwineAuthenticatesPypirc(t *testing.T) {
	feed := &protocol.Registry{Auth: protocol.AuthAzureArtifacts, PyPI: "https://pkgs.dev.azure.com/acme/web/_packaging/libs/pypi/simple/", PyPIUpload: "https://pkgs.dev.azure.com/acme/web/_packaging/libs/pypi/upload/"}
	t.Setenv("PYPIRC_PATH", "/agent/_temp/twineAuthenticate/.pypirc")
	got := targetOf(project(feed))
	if strings.Join(got.Args, " ") != "--config-file /agent/_temp/twineAuthenticate/.pypirc --repository libs" || got.Env != nil || got.Index != feed.PyPI {
		t.Errorf("in Azure Pipelines: %+v", got)
	}
	t.Setenv("PYPIRC_PATH", "")
	if got := targetOf(project(feed)); strings.Join(got.Args, " ") != "--repository-url "+feed.PyPIUpload {
		t.Errorf("on a laptop (artifacts-keyring signs in): %+v", got)
	}
}

func TestATokenRegistryUploadsAsTheAPITokenUser(t *testing.T) {
	t.Setenv("PY_FEED", "s3cr3t")
	got := targetOf(project(&protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "PY_FEED", PyPIUpload: "https://py.example.com/upload/"}))
	if got.Upload != "https://py.example.com/upload/" || got.Index != "https://py.example.com/simple/" || strings.Join(got.Env, " ") != "TWINE_USERNAME=__token__ TWINE_PASSWORD=s3cr3t" {
		t.Errorf("token: %+v", got)
	}
}

func TestWithoutARegistryItIsPyPIAsBefore(t *testing.T) {
	t.Setenv("TWINE_REPOSITORY_URL", "")
	t.Setenv("PYPI_TOKEN", "pypi-x")
	t.Setenv("TWINE_PASSWORD", "")
	got := targetOf(protocol.Project{Name: "sdk"})
	if got.Upload != "https://upload.pypi.org/legacy/" || got.Index != "https://pypi.org/simple/" || len(got.Args) != 0 || len(got.Env) != 2 {
		t.Errorf("pypi: %+v", got)
	}
}
