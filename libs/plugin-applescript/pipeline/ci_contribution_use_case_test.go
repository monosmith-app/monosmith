package pipeline

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	res := CI(protocol.CIParams{Projects: []protocol.Project{{}}})
	if len(res.Toolchains) != 0 || len(res.Setup) != 0 || len(res.Credentials) != 0 {
		t.Errorf("a macOS runner needs nothing installed: %+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := lookPath
	defer func() { lookPath = real }()
	projects := protocol.DoctorParams{Projects: []protocol.Project{{}}}

	lookPath = func(string) (string, error) { return "", errors.New("missing") }
	res := Doctor("", projects)
	if len(res.Findings) != 1 || res.Findings[0].OK || !res.Findings[0].Warning || res.Findings[0].Remedy == "" {
		t.Fatalf("without osacompile it is a warning with a remedy: %+v", res)
	}
	lookPath = func(string) (string, error) { return "/usr/bin/osacompile", nil }
	if res := Doctor("", projects); !res.Findings[0].OK {
		t.Errorf("%+v", res)
	}
	if res := Doctor("", protocol.DoctorParams{}); len(res.Findings) != 0 {
		t.Errorf("no projects, no findings: %+v", res)
	}
}
