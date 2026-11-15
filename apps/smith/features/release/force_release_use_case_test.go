package release

import (
	"strings"
	"testing"
)

func forceInfoFor(name string) (ForceInfo, bool) {
	switch name {
	case "mvd":
		return ForceInfo{Plugin: "node", TagPrefix: "mvd@", Latest: Version{Minor: 14}, Released: true, LatestTag: "mvd@0.14.0"}, true
	case "fresh":
		return ForceInfo{Plugin: "node", TagPrefix: "fresh@"}, true
	case "core":
		return ForceInfo{Plugin: "node", TagPrefix: "core@", Latest: Version{Major: 1, Minor: 2}, Released: true, LatestTag: "core@1.2.0"}, true
	}
	return ForceInfo{}, false
}

func TestForcePlanReleasesAPackageWithoutACommit(t *testing.T) {
	plan := Plan{Items: []Item{{Project: "web", Next: "1.0.1", Tag: "web@1.0.1"}}, Unchanged: []string{"mvd", "core"}}
	out, rest, notes, err := ForcePlan(plan, map[string]string{"mvd": "patch"}, []string{"mvd", "core", "web"}, forceInfoFor, nil)
	if err != nil {
		t.Fatal(err)
	}
	if names(out) != "mvd,web" || len(rest) != 0 || len(notes) != 1 {
		t.Fatalf("items %s rest %v notes %v", names(out), rest, notes)
	}
	item := out.Items[0]
	if !item.Manual || item.Next != "0.14.1" || item.Tag != "mvd@0.14.1" || item.Current != "0.14.0" || item.Level != "patch" {
		t.Fatalf("%+v", item)
	}
	if !strings.Contains(item.Reason, "no commit since mvd@0.14.0 releases this: you are releasing the same code under a new version") {
		t.Fatal(item.Reason)
	}
	if strings.Join(out.Unchanged, ",") != "core" {
		t.Fatal(out.Unchanged)
	}
	if TagMessage(item) == "mvd 0.14.1" || !strings.Contains(TagMessage(item), "manual release") {
		t.Fatal(TagMessage(item))
	}
	if !strings.Contains(Consequences(out, true), "mvd has no commit that releases it") {
		t.Fatal(Consequences(out, true))
	}
}

func TestForcePlanLeavesAPackageWithACommitToAdjust(t *testing.T) {
	plan := Plan{Items: []Item{{Project: "web", Next: "1.0.1"}}}
	_, rest, notes, err := ForcePlan(plan, map[string]string{"web": "minor", "ghost": "patch"}, []string{"web"}, forceInfoFor, nil)
	if err != nil || len(notes) != 0 || rest["web"] != "minor" || rest["ghost"] != "patch" {
		t.Fatalf("%v %v %v", rest, notes, err)
	}
}

func TestForcePlanOnAChannelUsesTheChannelCounter(t *testing.T) {
	ch := &Channel{ID: "beta", Major: 2, Tag: "beta"}
	plan := Plan{Channel: "beta", Unchanged: []string{"mvd"}}
	info := func(string) (ForceInfo, bool) {
		return ForceInfo{TagPrefix: "mvd@", Latest: Version{Major: 2, Pre: "beta.3"}, Released: true, OnChannel: true, LatestTag: "mvd@2.0.0-beta.3"}, true
	}
	out, _, _, err := ForcePlan(plan, map[string]string{"mvd": "patch"}, []string{"mvd"}, info, ch)
	if err != nil || out.Items[0].Next != "2.0.0-beta.4" || out.Items[0].Tag != "mvd@2.0.0-beta.4" {
		t.Fatalf("%+v %v", out.Items, err)
	}
}

func TestForcePlanRefuses(t *testing.T) {
	plan := Plan{Unchanged: []string{"fresh", "mvd"}}
	for name, bump := range map[string]string{"fresh": "patch", "mvd": "huge"} {
		if _, _, _, err := ForcePlan(plan, map[string]string{name: bump}, nil, forceInfoFor, nil); err == nil {
			t.Fatalf("%s=%s was accepted", name, bump)
		}
	}
}
