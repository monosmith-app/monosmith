package configedit

import "testing"

func TestDefaultAndExplanationSplitTheDescription(t *testing.T) {
	f := Field{Description: `Pipelines to generate. Default: ["github"].`}
	if got := f.Default(); got != `["github"]` {
		t.Fatalf("default %q", got)
	}
	if got := f.Explanation(); got != "Pipelines to generate." {
		t.Fatalf("explanation %q", got)
	}
	plain := Field{Description: "Just text."}
	if plain.Default() != "" || plain.Explanation() != "Just text." {
		t.Fatalf("a description without a default is all explanation: %q %q", plain.Default(), plain.Explanation())
	}
}
