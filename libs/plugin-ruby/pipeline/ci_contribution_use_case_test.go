package pipeline

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	ws := t.TempDir()
	res := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Kind: protocol.ProjectPackage}}})
	if res.Toolchains[0].Version != RubyVersion || res.Setup[0].Command[1] != "install" || res.Credentials[0].Env != "GEM_HOST_API_KEY" {
		t.Fatalf("%+v", res)
	}
	_ = os.WriteFile(filepath.Join(ws, ".ruby-version"), []byte("ruby-3.3.6\n"), 0o644)
	if v := Version(ws); v != "3.3.6" {
		t.Fatalf(".ruby-version pins %s", v)
	}
}

func TestCIAuditsTheBundleForKnownVulnerabilities(t *testing.T) {
	ws := t.TempDir()
	_ = os.MkdirAll(filepath.Join(ws, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(ws, "apps", "web", "web.gemspec"), nil, 0o644)
	got := CI(ws, protocol.CIParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}}}).Audit
	if len(got) != 1 || got[0].Blocking {
		t.Fatalf("one report-only step: %+v", got)
	}
	if want := []string{"Gemfile", "Gemfile.lock", "apps/web/web.gemspec"}; !reflect.DeepEqual(got[0].CacheFiles, want) {
		t.Errorf("the verdict is remembered per Gemfile, lockfile and gemspec: %v", got[0].CacheFiles)
	}
	script := got[0].Command[2]
	for _, want := range []string{"bundle-audit check --update", "gem install bundler-audit", `"$CI"`} {
		if !strings.Contains(script, want) {
			t.Errorf("the step says how to get the tool (%q missing): %s", want, script)
		}
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "bundle" {
			return "", errors.New("missing")
		}
		return "", nil
	}
	res := Doctor(t.TempDir(), protocol.DoctorParams{Projects: []protocol.Project{{}}})
	if len(res.Findings) != 2 || !res.Findings[0].OK || res.Findings[1].OK {
		t.Fatalf("%+v", res)
	}
}
