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
	res := CI(protocol.CIParams{Projects: []protocol.Project{{}}})
	if len(res.Toolchains) != 1 || res.Toolchains[0].GitHub.Uses == "" || res.Toolchains[0].Install == "" {
		t.Errorf("%+v", res)
	}
}

func TestDoctor(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	projects := protocol.DoctorParams{Projects: []protocol.Project{{}}}

	probe = func(string, []string) (string, error) { return "", errors.New("missing") }
	if res := Doctor(t.TempDir(), projects); len(res.Findings) != 1 || res.Findings[0].OK || res.Findings[0].Remedy == "" {
		t.Fatalf("%+v", res)
	}
	probe = func(string, []string) (string, error) { return "Swift version 6.0", nil }
	if res := Doctor(t.TempDir(), projects); !res.Findings[0].OK {
		t.Errorf("%+v", res)
	}
	if res := Doctor(t.TempDir(), protocol.DoctorParams{}); len(res.Findings) != 0 {
		t.Errorf("no projects, no findings: %+v", res)
	}
}
