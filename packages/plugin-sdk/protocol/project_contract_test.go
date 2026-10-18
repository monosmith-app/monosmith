package protocol

import "testing"

func TestSetToolchainProbesOnceAndOnlyForProjects(t *testing.T) {
	probes := 0
	probe := func() string { probes++; return "node 20.11.0" }

	var none DiscoverResult
	none.SetToolchain(probe)
	if probes != 0 {
		t.Fatal("no project, no probe")
	}

	r := DiscoverResult{Projects: []Project{{Name: "a"}, {Name: "b"}}}
	r.SetToolchain(probe)
	if probes != 1 || r.Projects[0].Toolchain != "node 20.11.0" || r.Projects[1].Toolchain != "node 20.11.0" {
		t.Errorf("%d probes, %+v", probes, r.Projects)
	}
}
