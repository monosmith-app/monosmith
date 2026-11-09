package releasepack

import (
	"strings"
	"testing"
)

func TestTheBundleNamesEveryPackageAndInstallsThemTogether(t *testing.T) {
	b := NewBundle("o/r", []Package{npmPackage("cli", "0.2.4", "cli-0.2.4.tgz"), npmPackage("core", "0.5.2", "core-0.5.2.tgz")})
	if len(b.Install) != 1 || b.Install[0].Ecosystem != "npm" {
		t.Fatalf("one install per ecosystem: %+v", b.Install)
	}
	if b.Install[0].Command != "npm i ./cli-0.2.4.tgz ./core-0.5.2.tgz" {
		t.Fatalf("command %q", b.Install[0].Command)
	}
	if len(b.Install[0].Download) != 2 || b.Install[0].Download[1] != `gh release download "core@0.5.2" --repo o/r --pattern '*.tgz' --pattern '*.sha256' --dir .` {
		t.Fatalf("one download per tag: %q", b.Install[0].Download)
	}
	text := BundleInstructions(b)
	for _, want := range []string{"gh auth login", "npm i ./cli-0.2.4.tgz ./core-0.5.2.tgz", "sha256sum -c *.sha256", strings.Repeat("a", 64)} {
		if !strings.Contains(text, want) {
			t.Errorf("INSTALL.md lacks %q", want)
		}
	}
}

func TestEachEcosystemGetsItsOwnInstallCommand(t *testing.T) {
	b := NewBundle("o/r", []Package{
		{Project: "py", Ecosystem: "Python", Version: "1.0.0", Tag: "py@1.0.0", Files: []File{{Name: "acme-1.0.0-py3-none-any.whl"}, {Name: "acme-1.0.0.tar.gz"}}},
		{Project: "net", Ecosystem: "NuGet", Version: "1.0.0", Tag: "net@1.0.0", Files: []File{{Name: "Acme.Core.1.0.0.nupkg"}}},
		{Project: "rb", Ecosystem: "RubyGems", Version: "1.0.0", Tag: "rb@1.0.0", Files: []File{{Name: "acme-1.0.0.gem"}}},
		{Project: "rs", Ecosystem: "crates", Version: "1.0.0", Tag: "rs@1.0.0", Files: []File{{Name: "acme-1.0.0.crate"}}},
	})
	got := map[string]string{}
	for _, i := range b.Install {
		got[i.Ecosystem] = i.Command
	}
	if got["Python"] != "pip install ./acme-1.0.0-py3-none-any.whl" || got["NuGet"] != "dotnet add package Acme.Core --version 1.0.0 --source ." || got["RubyGems"] != "gem install ./acme-1.0.0.gem" {
		t.Fatalf("%v", got)
	}
	if !strings.Contains(got["crates"], "tar -xzf ./acme-1.0.0.crate") || !strings.Contains(got["crates"], `acme = { path = "acme-1.0.0" }`) {
		t.Fatalf("crates: %q", got["crates"])
	}
}
