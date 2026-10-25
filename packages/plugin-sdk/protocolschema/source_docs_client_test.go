package protocolschema

import "testing"

func TestSourceDocsReadsCommentsAndConstants(t *testing.T) {
	docs := docsFor(t, `package protocol

// Widget is a thing.
//
// It has two paragraphs.
type Widget struct {
	// Size is how big.
	Size int `+"`json:\"size\"`"+`
	Weight int `+"`json:\"weight\"`"+` // in grams
}

type Colour string

const (
	// Red is warm.
	Red Colour = "red"
	Blue Colour = "blue"
)

const (
	CodeBad  = -32700
	CodeFine = 7
	Other    = 1 + 2 // not a literal
)
`)
	if got := docs.TypeDoc("Widget"); got != "Widget is a thing.\n\nIt has two paragraphs." {
		t.Errorf("type doc %q", got)
	}
	if docs.FieldDoc("Widget", "Size") != "Size is how big." || docs.FieldDoc("Widget", "Weight") != "in grams" {
		t.Errorf("field docs %q %q", docs.FieldDoc("Widget", "Size"), docs.FieldDoc("Widget", "Weight"))
	}
	enum := docs.Enum("Colour")
	if len(enum) != 2 || enum[0].Value != "red" || enum[0].Doc != "Is warm." || enum[1].Value != "blue" {
		t.Errorf("enum %+v", enum)
	}
	codes := docs.ConstsWithPrefix("Code")
	if len(codes) != 2 || codes[0].Name != "CodeBad" || codes[0].Value != "-32700" || codes[1].Value != "7" {
		t.Errorf("codes %+v", codes)
	}
	if _, ok := docs.Const("Other"); ok {
		t.Error("a constant that is not a literal was read")
	}
}
