package pipeline

import "testing"

func TestDropUse(t *testing.T) {
	src := "go 1.26.0\n\nuse (\n\t./apps/api\n\t\"./libs/core\" // shared\n\t./libs/kit\n)\n\nuse ./packages/sdk\n"
	out, ok := DropUse(src, "libs/core")
	if !ok || out != "go 1.26.0\n\nuse (\n\t./apps/api\n\t./libs/kit\n)\n\nuse ./packages/sdk\n" {
		t.Fatalf("%v\n%s", ok, out)
	}
	out, ok = DropUse(out, "packages/sdk")
	if !ok || out != "go 1.26.0\n\nuse (\n\t./apps/api\n\t./libs/kit\n)\n\n" {
		t.Fatalf("%v\n%q", ok, out)
	}
	if _, ok := DropUse(out, "libs/missing"); ok {
		t.Fatal("nothing to drop")
	}
}
