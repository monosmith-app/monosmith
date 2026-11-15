package release

import "testing"

func TestRenderNotes(t *testing.T) {
	item := Item{Project: "web", Next: "2.0.0", Changes: []Change{
		{Type: "feat", Scopes: []string{"auth"}, Description: "sign in with passkeys", Hash: "aaaaaaa111", Author: "Ada"},
		{Type: "fix", Description: "keep the session on reload", Hash: "bbbbbbb222", Author: "Grace"},
		{Type: "feat", Description: "drop the v1 API", Breaking: true, BreakingNote: "Clients must call /v2.\nSee the migration guide.", Hash: "ccccccc333", Author: "Ada"},
		{Type: "perf", Description: "cache the menu", Hash: "ddddddd444", Author: "dependabot[bot]"},
	}}
	want := "## ▲ Breaking changes\n\n" +
		"- drop the v1 API (ccccccc)\n  Clients must call /v2.\n  See the migration guide.\n\n" +
		"## Features\n\n- **auth:** sign in with passkeys (aaaaaaa)\n- drop the v1 API (ccccccc)\n\n" +
		"## Fixes\n\n- keep the session on reload (bbbbbbb)\n\n" +
		"## Performance\n\n- cache the menu (ddddddd)\n\n" +
		"## Thanks\n\nThanks to Ada and Grace.\n"
	if got := RenderNotes(item); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if got := RenderNotes(Item{Reason: "dependency ui releases"}); got != "## Dependencies\n\n- dependency ui releases\n" {
		t.Fatalf("a dependency release:\n%s", got)
	}
}

func TestBreakingNote(t *testing.T) {
	c := ParseCommit("h", "feat!: new config", "Body text.\n\nBREAKING CHANGE: the old keys are gone.\nRename them.\n\nRefs: #12")
	if !c.Breaking || c.BreakingNote != "the old keys are gone.\nRename them." {
		t.Fatalf("%+v", c)
	}
}
