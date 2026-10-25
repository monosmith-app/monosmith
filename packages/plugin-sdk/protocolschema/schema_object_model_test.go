package protocolschema

import (
	"strings"
	"testing"
)

func TestRenderKeepsOrderAndInlinesShortScalars(t *testing.T) {
	doc := Object{
		{Key: "z", Value: Object{{Key: "type", Value: "string"}}},
		{Key: "a", Value: []string{"x", "y"}},
		{Key: "n", Value: nil},
		{Key: "long", Value: Object{{Key: "description", Value: strings.Repeat("w", 120)}}},
		{Key: "html", Value: "<a & b>"},
	}
	got, err := Render(doc)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"z\": {\"type\": \"string\"},\n  \"a\": [\"x\", \"y\"],\n  \"n\": null,\n  \"long\": {\n    \"description\": \"" +
		strings.Repeat("w", 120) + "\"\n  },\n  \"html\": \"<a & b>\"\n}\n"
	if string(got) != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestSetReplacesInPlace(t *testing.T) {
	o := Object{{Key: "a", Value: 1}, {Key: "b", Value: 2}}
	o.Set("a", 3)
	o.Set("c", 4)
	if v, _ := o.Get("a"); v != 3 || len(o) != 3 || o[0].Key != "a" || o[2].Key != "c" {
		t.Fatalf("%v", o)
	}
}

func TestRenderRefusesWhatItCannotWrite(t *testing.T) {
	if _, err := Render(Object{{Key: "x", Value: struct{}{}}}); err == nil {
		t.Fatal("a struct value was rendered")
	}
}
