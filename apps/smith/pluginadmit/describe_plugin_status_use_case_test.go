package pluginadmit

import (
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
)

func TestDescribeShowsNamesNeverValues(t *testing.T) {
	ws, p, _ := setup(t)
	p.WorkspaceEnv = []string{"NPM_TOKEN", "SMITH_CACHE_KEY"}
	p.ManifestEnv = []string{"AWS_*", "SMITH_CACHE_TOKEN"}
	p.Pin = "abc"
	store, err := plugintrust.OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	st := Describe(store, ws, []Plugin{p})[0]
	want := []string{"AWS_*", "NPM_TOKEN", "SMITH_CACHE_KEY"}
	if !slices.Equal(st.Env, want) {
		t.Fatalf("env %v, want %v (a manifest cannot ask for smith's secrets; smith.toml can)", st.Env, want)
	}
	if st.Trusted || !st.Pinned || st.Reason != plugintrust.ReasonNew {
		t.Fatalf("status %+v", st)
	}
	if err := store.Trust(ws, "widget", st.Fingerprint); err != nil {
		t.Fatal(err)
	}
	if got := Describe(store, ws, []Plugin{p})[0]; !got.Trusted {
		t.Fatalf("not trusted after trusting: %+v", got)
	}
}
