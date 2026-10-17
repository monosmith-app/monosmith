package convertplan

import (
	"fmt"
	"testing"
)

// officeReaderRollup is OpenCraw's office-reader config before converting:
// Rollup's own input map, beside comments whose apostrophes and braces must
// not confuse the reader.
const officeReaderRollup = "const { withNx } = require('@nx/rollup/with-nx')\n" +
	"module.exports = withNx(\n  {\n    main:       './src/index.ts',\n    outputPath: './dist',\n    tsConfig:   './tsconfig.lib.json',\n" +
	"    // Swapped from swc by MoNecromanCI. @nx/rollup runs swc without {sourceMaps}\n    format:     ['esm'],\n  },\n  {\n" +
	"    // One entry per format, so `@opencraw/office-reader/xlsx` loads the spreadsheet reader alone.\n" +
	"    input: {\n" +
	"      index: `${__dirname}/src/index.ts`,\n" +
	"      xlsx:  `${__dirname}/src/spreadsheet/index.ts`,\n" +
	"      pptx:  path.join(__dirname, 'src/presentation/index.ts'),\n" +
	"      'docx': './src/document/index.ts',\n" +
	"    },\n" +
	"    output: { sourcemapPathTransform: p => p.replaceAll(String.fromCodePoint(92), '/') },\n  },\n)\n"

func TestRollupEntriesReadsAnInputMap(t *testing.T) {
	entries, ok := RollupEntries(officeReaderRollup)
	want := "[{index src/index.ts} {xlsx src/spreadsheet/index.ts} {pptx src/presentation/index.ts} {docx src/document/index.ts}]"
	if !ok || fmt.Sprint(entries) != want {
		t.Fatalf("entries %v %v", entries, ok)
	}
}

func TestRollupEntriesNamesListedEntriesAfterTheirFiles(t *testing.T) {
	for _, text := range []string{
		"withNx({ main: './src/index.ts' }, { input: ['./src/index.ts', './src/cli.ts'] })",
		"withNx({ main: './src/index.ts', additionalEntryPoints: ['./src/index.ts', `${__dirname}/src/cli.ts`] })",
	} {
		entries, ok := RollupEntries(text)
		if !ok || fmt.Sprint(entries) != "[{index src/index.ts} {cli src/cli.ts}]" {
			t.Errorf("%s: %v %v", text, entries, ok)
		}
	}
}

func TestRollupEntriesRefusesWhatItCannotRead(t *testing.T) {
	for _, text := range []string{
		"withNx({}, { input: entries })",
		"withNx({}, { input: { index: resolveEntry('index') } })",
		"withNx({}, { input: ['./src/a/index.ts', './src/b/index.ts'] })",
		"withNx({}, { input: { up: '../other/index.ts' } })",
	} {
		if entries, ok := RollupEntries(text); ok {
			t.Errorf("%s read as %v", text, entries)
		}
	}
	if entries, ok := RollupEntries("withNx({ main: './src/index.ts' })"); !ok || entries != nil {
		t.Errorf("main alone: %v %v", entries, ok)
	}
}

func TestRollupMultiEntryBuildsEveryEntryInOneRun(t *testing.T) {
	entries, _ := RollupEntries(officeReaderRollup)
	got, ok := rollupMultiEntry(entries, "esm", "dist/index.esm.js")
	want := "esbuild index=src/index.ts xlsx=src/spreadsheet/index.ts pptx=src/presentation/index.ts docx=src/document/index.ts --bundle --splitting --packages=external --format=esm --outdir=dist --out-extension:.js=.esm.js --chunk-names=chunks/[name]-[hash] --sourcemap"
	if !ok || got != want {
		t.Fatalf("got %q", got)
	}
	if _, ok := rollupMultiEntry(entries, "cjs", "dist/index.cjs.js"); ok {
		t.Error("esbuild cannot split CommonJS")
	}
	if _, ok := rollupMultiEntry(entries, "esm", "dist/main.esm.js"); ok {
		t.Error("main is none of the entries")
	}
}
