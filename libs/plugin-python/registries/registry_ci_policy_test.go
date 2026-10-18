package registries

import (
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAzureFeedNamesTheFeedAsAzureTasksDo(t *testing.T) {
	cases := map[string]string{
		"https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/":       "libs",
		"https://pkgs.dev.azure.com/acme/web/_packaging/libs/pypi/upload/":   "web/libs",
		"https://acme.pkgs.visualstudio.com/_packaging/libs/pypi/simple/":    "libs",
		"https://acme.pkgs.visualstudio.com/web/_packaging/libs/pypi/simple": "web/libs",
		"https://pypi.example.com/simple/":                                   "",
	}
	for in, want := range cases {
		if got := AzureFeed(in); got != want {
			t.Errorf("%s: %q, want %q", in, got, want)
		}
	}
}

func TestCISignsInToAzureArtifactsWithoutASecret(t *testing.T) {
	feed := &protocol.Registry{Auth: protocol.AuthAzureArtifacts, PyPI: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/", PyPIUpload: "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/upload/"}
	r, public := CI([]protocol.Project{{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: feed, Publish: feed}}})
	if public || len(r.Credentials) != 0 || len(r.Authenticate) != 2 {
		t.Fatalf("result %+v public=%v", r, public)
	}
	if a := r.Authenticate[0].Azure; a.Uses != "PipAuthenticate@1" || a.Inputs[0].Value != "libs" {
		t.Errorf("pip: %+v", a)
	}
	if a := r.Authenticate[1].Azure; a.Uses != "TwineAuthenticate@1" || a.Inputs[0].Value != "libs" {
		t.Errorf("twine: %+v", a)
	}
}

func TestCIPassesATokenRegistrysVariable(t *testing.T) {
	tok := &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "PY_FEED", PyPI: "https://p/simple/", PyPIUpload: "https://p/upload/"}
	r, public := CI([]protocol.Project{{Kind: protocol.ProjectPackage, Registries: &protocol.Registries{Install: tok, Publish: tok}}})
	if public || len(r.Credentials) != 1 || r.Credentials[0].Env != "PY_FEED" || !r.Credentials[0].Setup {
		t.Errorf("result %+v public=%v", r, public)
	}
	if _, public := CI([]protocol.Project{{Kind: protocol.ProjectPackage}}); !public {
		t.Error("no registry: published to PyPI")
	}
}
