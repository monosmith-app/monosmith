package derive

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func keysFile(renderer string) protocol.DerivedFile {
	return protocol.DerivedFile{Path: "a/tsconfig.json", Mode: protocol.DeriveKeys, Format: "json", Renderer: renderer, Version: 1, Keys: map[string]any{"compilerOptions": map[string]any{}}}
}

func TestTheWholeFileRendererOwnsAFileThatAKeysRendererAlsoClaims(t *testing.T) {
	got := oneOwnerPerFile(derivelock.Record{}, []Proposal{
		{Plugin: "node", File: keysFile("node.sibling-shape")},
		{Plugin: "typescript", File: wholeFile("typescript.tsconfig", "{}\n")},
	})
	if len(got) != 1 || got[0].File.Renderer != "typescript.tsconfig" {
		t.Fatalf("%+v", got)
	}
}

func TestTheRecordedRendererOwnsAFileThatTwoKindsOfRendererClaim(t *testing.T) {
	rec := derivelock.Record{Files: []derivelock.Entry{{Path: "a/tsconfig.json", Owner: derivelock.OwnerSmith, Renderer: "node.sibling-shape", Mode: protocol.DeriveKeys}}}
	got := oneOwnerPerFile(rec, []Proposal{
		{Plugin: "node", File: keysFile("node.sibling-shape")},
		{Plugin: "typescript", File: wholeFile("typescript.tsconfig", "{}\n")},
	})
	if len(got) != 1 || got[0].File.Renderer != "node.sibling-shape" {
		t.Fatalf("%+v", got)
	}
}

func TestRenderersOfDifferentKeysKeepTheirKeys(t *testing.T) {
	in := []Proposal{{Plugin: "a", File: keysFile("one")}, {Plugin: "b", File: keysFile("two")}}
	if got := oneOwnerPerFile(derivelock.Record{}, in); len(got) != 2 {
		t.Fatalf("%+v", got)
	}
}

// A file that this run creates counts as owned in a second answer, so a file
// that depends on it is right in the first run.
func TestOneRunSeesTheFileThatTheRunCreates(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	calls := 0
	collect := func(owned []string) ([]Proposal, []Divergence, error) {
		calls++
		files := []protocol.DerivedFile{{Path: "a/tsconfig.lib.json", Mode: protocol.DeriveFile, Renderer: "r", Content: "{}\n"}}
		if slices.Contains(owned, "a/tsconfig.lib.json") {
			files = append(files, protocol.DerivedFile{Path: "b/tsconfig.json", Mode: protocol.DeriveFile, Renderer: "r", Content: "refs lib\n"})
		}
		var out []Proposal
		for _, f := range files {
			out = append(out, Proposal{Plugin: "p", File: f})
		}
		return out, nil, nil
	}
	got, _, err := converge(root, nil, collect)
	if err != nil || len(got) != 2 || calls != 2 {
		t.Fatalf("err %v, %d proposals, %d calls", err, len(got), calls)
	}
}

func TestNoSecondAnswerWhenNothingIsNew(t *testing.T) {
	calls := 0
	_, _, err := converge(t.TempDir(), []string{"a/x.json"}, func([]string) ([]Proposal, []Divergence, error) {
		calls++
		return []Proposal{{File: protocol.DerivedFile{Path: "a/x.json"}}}, nil, nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("err %v, %d calls", err, calls)
	}
}

func TestConvergeStopsOnAPluginError(t *testing.T) {
	_, _, err := converge(t.TempDir(), nil, func([]string) ([]Proposal, []Divergence, error) {
		return nil, nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("no error")
	}
}
