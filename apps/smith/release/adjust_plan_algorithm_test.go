package release

import (
	"bufio"
	"strings"
	"testing"
)

func samplePlan() Plan {
	return Plan{
		Items: []Item{
			{Project: "core", Current: "1.2.0", Next: "1.3.0", Level: "minor", Tag: "core@1.3.0", Reason: "2 releasing commit(s)"},
			{Project: "api", Current: "2.0.1", Next: "2.0.2", Level: "patch", Tag: "api@2.0.2", Reason: "dependency core 1.3.0"},
			{Project: "web", Current: "0.4.0", Next: "0.4.1", Level: "patch", Tag: "web@0.4.1", Reason: "1 releasing commit(s)"},
			{Project: "fresh", Next: "0.1.0", Level: "minor", Tag: "fresh@0.1.0", Reason: "1 releasing commit(s)"},
			{Project: "pinned", Current: "1.0.0", Next: "1.5.0", Level: "forced", Tag: "pinned@1.5.0", Forced: true},
		},
		Unchanged: []string{"docs"},
	}
}

func names(p Plan) string {
	var n []string
	for _, i := range p.Items {
		n = append(n, i.Project)
	}
	return strings.Join(n, ",")
}

func TestAdjust(t *testing.T) {
	for _, tc := range []struct {
		name    string
		a       Adjustments
		wantErr string
		check   func(t *testing.T, p Plan)
	}{
		{name: "no change keeps the plan", check: func(t *testing.T, p Plan) {
			if names(p) != "core,api,web,fresh,pinned" {
				t.Fatal(names(p))
			}
		}},
		{name: "hold removes the package", a: Adjustments{Hold: []string{"web"}}, check: func(t *testing.T, p Plan) {
			if names(p) != "core,api,fresh,pinned" {
				t.Fatal(names(p))
			}
		}},
		{name: "hold cascades to a package that releases only for it", a: Adjustments{Hold: []string{"core"}}, check: func(t *testing.T, p Plan) {
			if names(p) != "web,fresh,pinned" {
				t.Fatal(names(p))
			}
		}},
		{name: "hold of an unchanged package is a no-op", a: Adjustments{Hold: []string{"docs"}}, check: func(t *testing.T, p Plan) {
			if len(p.Items) != 5 {
				t.Fatal(names(p))
			}
		}},
		{name: "hold of an unknown package", a: Adjustments{Hold: []string{"nope"}}, wantErr: "--hold nope"},
		{name: "bump major recomputes next and tag", a: Adjustments{Bump: map[string]string{"web": "major"}}, check: func(t *testing.T, p Plan) {
			var web Item
			for _, i := range p.Items {
				if i.Project == "web" {
					web = i
				}
			}
			// Before 1.0 a major bump is a minor one.
			if web.Next != "0.5.0" || web.Tag != "web@0.5.0" || web.Level != "major" || !strings.Contains(web.Reason, "you chose major") {
				t.Fatalf("%+v", web)
			}
		}},
		{name: "bump patch of a minor", a: Adjustments{Bump: map[string]string{"core": "patch"}}, check: func(t *testing.T, p Plan) {
			if p.Items[0].Next != "1.2.1" || p.Items[0].Tag != "core@1.2.1" {
				t.Fatalf("%+v", p.Items[0])
			}
		}},
		{name: "bump of an unreleased package", a: Adjustments{Bump: map[string]string{"fresh": "major"}}, wantErr: "never released"},
		{name: "bump of a forced version", a: Adjustments{Bump: map[string]string{"pinned": "major"}}, wantErr: "already sets the version"},
		{name: "bump with an unknown level", a: Adjustments{Bump: map[string]string{"web": "huge"}}, wantErr: "patch, minor or major"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := Adjust(samplePlan(), tc.a)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, got)
		})
	}
}

func TestAdjustRefusesABumpOnAChannel(t *testing.T) {
	p := samplePlan()
	p.Channel = "beta"
	if _, _, err := Adjust(p, Adjustments{Bump: map[string]string{"web": "minor"}}); err == nil {
		t.Fatal("want a refusal")
	}
}

func TestParseBumps(t *testing.T) {
	got, err := ParseBumps([]string{"web=minor", "api=major"})
	if err != nil || got["web"] != "minor" || got["api"] != "major" {
		t.Fatal(got, err)
	}
	for _, bad := range []string{"web", "=minor", "web=none", "web=big"} {
		if _, err := ParseBumps([]string{bad}); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}

func TestConsequencesNameTagsAndIrreversibility(t *testing.T) {
	s := Consequences(samplePlan(), true)
	for _, want := range []string{"cannot be undone", "core@1.3.0", "push them to the remote", "registries"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in %s", want, s)
		}
	}
	if !strings.Contains(Consequences(samplePlan(), false), "--no-push") {
		t.Error("--no-push should be named")
	}
}

func TestReviewLetsThePersonHoldAndBump(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
		ok                bool
	}{
		{"enter releases all", "\n", "core,api,web,fresh,pinned", true},
		{"cancel", "4\n", "", false},
		{"hold web then release", "3\n3\n\n", "core,api,fresh,pinned", true},
		{"bump web to minor then release", "2\n3\n2\n\n", "core,api,web,fresh,pinned", true},
		{"closed input releases nothing it was not asked", "", "core,api,web,fresh,pinned", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			got, ok, err := Review(bufio.NewReader(strings.NewReader(tc.input)), &out, samplePlan(), func(Plan) {}, Adjustments{})
			if err != nil {
				t.Fatal(err)
			}
			if ok != tc.ok || (tc.ok && names(got) != tc.want) {
				t.Fatalf("ok=%v plan=%s\n%s", ok, names(got), out.String())
			}
		})
	}
}
