package pipeline

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCI(t *testing.T) {
	if res := CI(protocol.CIParams{}); len(res.Toolchains) != 0 {
		t.Error("no projects, no toolchain")
	}
	if res := CI(protocol.CIParams{Projects: []protocol.Project{{}}}); len(res.Toolchains) != 1 || res.Toolchains[0].Install == "" {
		t.Errorf("%+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := lookPath
	defer func() { lookPath = real }()
	projects := protocol.DoctorParams{Projects: []protocol.Project{{}}}

	lookPath = func(string) (string, error) { return "", errors.New("missing") }
	res := Doctor("", projects)
	if len(res.Findings) != 2 || res.Findings[0].OK || res.Findings[1].OK || !res.Findings[0].Warning || res.Findings[1].Remedy == "" {
		t.Fatalf("without the tools both are warnings with a remedy: %+v", res)
	}
	lookPath = func(name string) (string, error) {
		if name == "pod" {
			return "/usr/bin/pod", nil
		}
		return "", errors.New("missing")
	}
	if res := Doctor("", projects); !res.Findings[0].OK || res.Findings[1].OK {
		t.Errorf("%+v", res)
	}
	if res := Doctor("", protocol.DoctorParams{}); len(res.Findings) != 0 {
		t.Errorf("no projects, no findings: %+v", res)
	}
}
