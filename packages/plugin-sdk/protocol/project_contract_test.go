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

func TestForReleaseTakesTheReleaseProfile(t *testing.T) {
	debug := Task{Command: []string{"cargo", "build"}, Ship: []string{"debug/api"}, Outputs: []string{"out"}}
	if got := debug.ForRelease(); len(got.Command) != 2 || got.Ship[0] != "debug/api" {
		t.Errorf("a task without a release profile changed: %+v", got)
	}
	both := debug
	both.Release = &TaskProfile{Command: []string{"cargo", "build", "--release"}, Ship: []string{"release/api"}}
	got := both.ForRelease()
	if len(got.Command) != 3 || got.Ship[0] != "release/api" || got.Outputs[0] != "out" {
		t.Errorf("release profile: %+v", got)
	}
	if len(both.Command) != 2 {
		t.Errorf("ForRelease changed the task it was called on: %+v", both.Command)
	}
}
