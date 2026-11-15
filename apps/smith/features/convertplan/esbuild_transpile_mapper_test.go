package convertplan

import (
	"strings"
	"testing"
)

func TestEsbuildTranspileReadsTheTsconfigAndBundlesNothing(t *testing.T) {
	line, ok := EsbuildTranspile("tsconfig.app.json", "src/main.ts", "dist", "node", "esm")
	if !ok {
		t.Fatal("not converted")
	}
	for _, want := range []string{
		`node -e "`,
		`ts.getParsedCommandLineOfConfigFile('tsconfig.app.json'`,
		`entryPoints:[...new Set(['src/main.ts',...c.fileNames].map(f=>p.resolve(f)))]`,
		`outdir:'dist',platform:'node',format:'esm',sourcemap:true,tsconfig:'tsconfig.app.json'`,
	} {
		if !strings.Contains(line, want) {
			t.Errorf("missing %q in\n%s", want, line)
		}
	}
	if strings.Contains(line, "bundle") {
		t.Errorf("bundles: %s", line)
	}
	if _, ok := EsbuildTranspile("tsconfig.app.json", "src/$main.ts", "dist", "node", "esm"); ok {
		t.Error("a path a shell would read is converted")
	}
	if _, ok := EsbuildTranspile("", "src/main.ts", "dist", "node", "esm"); ok {
		t.Error("converted without a tsconfig")
	}
}
