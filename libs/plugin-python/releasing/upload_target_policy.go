package releasing

import (
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-python/registries"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// uploadTarget is where twine uploads a project and how it signs in.
type uploadTarget struct {
	// Upload is the upload URL, Index the simple index that shows what the
	// registry has.
	Upload, Index string
	// Args are twine's repository arguments; Env its credentials.
	Args []string
	Env  []string
}

// targetOf is the project's registry from smith.toml when it names one:
//   - Azure Artifacts: in Azure Pipelines, TwineAuthenticate@1 wrote a
//     .pypirc (PYPIRC_PATH) holding the pipeline's token under the feed's
//     name; elsewhere twine signs in through artifacts-keyring
//   - token: twine's API-token user with the variable smith.toml names
//
// Otherwise PyPI, as before: TWINE_REPOSITORY_URL picks the index and
// PYPI_TOKEN, when set, is the API token.
func targetOf(p protocol.Project) uploadTarget {
	if r := p.Registries; r != nil && r.Publish != nil && r.Publish.PyPIUpload != "" {
		reg := r.Publish
		t := uploadTarget{Upload: reg.PyPIUpload, Index: reg.PyPI, Args: []string{"--repository-url", reg.PyPIUpload}}
		if t.Index == "" {
			t.Index = SimpleIndexURL(reg.PyPIUpload)
		}
		switch reg.Auth {
		case protocol.AuthAzureArtifacts:
			feed := registries.AzureFeed(reg.PyPIUpload)
			if rc := os.Getenv("PYPIRC_PATH"); rc != "" && feed != "" {
				// TwineAuthenticate names the .pypirc section after the feed
				// alone, without its project.
				t.Args = []string{"--config-file", rc, "--repository", feed[strings.LastIndex(feed, "/")+1:]}
			}
		case protocol.AuthToken:
			if token := os.Getenv(reg.TokenEnv); token != "" {
				t.Env = []string{"TWINE_USERNAME=__token__", "TWINE_PASSWORD=" + token}
			}
		}
		return t
	}
	upload := os.Getenv("TWINE_REPOSITORY_URL")
	if upload == "" {
		upload = "https://upload.pypi.org/legacy/"
	}
	t := uploadTarget{Upload: upload, Index: SimpleIndexURL(upload)}
	if token := os.Getenv("PYPI_TOKEN"); token != "" && os.Getenv("TWINE_PASSWORD") == "" {
		t.Env = []string{"TWINE_USERNAME=__token__", "TWINE_PASSWORD=" + token}
	}
	return t
}
