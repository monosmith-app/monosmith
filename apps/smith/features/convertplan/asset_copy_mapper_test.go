package convertplan

import (
	"strings"
	"testing"
)

func TestAssetCopyCopiesEachEntryRelativeToTheProject(t *testing.T) {
	build := target(t, `{"executor": "@nx/esbuild:esbuild", "options": {"outputPath": "apps/ext/dist", "assets": [
		{"input": "apps/ext/src/studio-host", "glob": "studio-host.process.mjs", "output": "."},
		{"input": "libs/media", "glob": "**/*", "output": "media"},
		"apps/ext/src/assets"
	]}}`)
	line, ok := AssetCopy("apps/ext", build, "dist")
	if !ok {
		t.Fatal("not converted")
	}
	want := `[['src/studio-host','studio-host.process.mjs','dist',0],['../../libs/media','**/*','dist/media',0],['.','src/assets','dist',1]]`
	if !strings.HasPrefix(line, `node -e "`) || !strings.Contains(line, "for(const[i,g,o,w]of"+want+")") || !strings.HasSuffix(line, `"`) {
		t.Fatalf("line %s", line)
	}
}

func TestAssetCopyLeavesWhatItCannotCopy(t *testing.T) {
	for _, raw := range []string{
		`{"options": {"assets": [{"input": "a", "glob": "**/*", "output": ".", "ignore": ["*.map"]}]}}`,
		`{"options": {"assets": ["a/it's"]}}`,
		`{"options": {}}`,
	} {
		if line, ok := AssetCopy("apps/x", target(t, raw), "dist"); ok {
			t.Errorf("%s: converted to %s", raw, line)
		}
	}
}
