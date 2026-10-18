package tooling

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestStarterSpecTakesTheWorkspaceSuffix(t *testing.T) {
	write := func(root, rel string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte("x\n"), 0o644)
	}
	params := protocol.GenerateParams{Directory: "packages/driver"}
	for _, c := range []struct {
		existing, want string
	}{
		{"", "src/greeting/greet.use-case.spec.ts"},
		{"packages/core/src/run.use-case.test.ts", "src/greeting/greet.use-case.test.ts"},
		{"packages/core/src/run.use-case.spec.ts", "src/greeting/greet.use-case.spec.ts"},
	} {
		root := t.TempDir()
		write(root, "packages/driver/"+starterUseCase)
		if c.existing != "" {
			write(root, c.existing)
		}
		files := StarterSpec(root, params, "vitest")
		if _, ok := files[c.want]; !ok || len(files) != 1 {
			t.Errorf("with %q: files %v, want %s", c.existing, files, c.want)
		}
	}
}
