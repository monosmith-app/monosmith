package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func writeSource(t *testing.T, path, body string) {
	t.Helper()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Error(err)
	}
}

func TestScanImportsIgnoresAFileThatVanishedBeforeItWasRead(t *testing.T) {
	ws := t.TempDir()
	writeSource(t, filepath.Join(ws, "a.ts"), "import x from 'lodash'\n")
	if err := os.Symlink(filepath.Join(ws, "gone.ts"), filepath.Join(ws, "b.ts")); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	got, err := ScanImports(ws)
	if err != nil || !reflect.DeepEqual(got, []string{"lodash"}) {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestScanImportsSurvivesFilesDeletedDuringTheWalk(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows reports a file being deleted as access denied, not as missing")
	}
	ws := t.TempDir()
	writeSource(t, filepath.Join(ws, "keep.ts"), "import x from 'lodash'\n")
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			for j := 0; j < 20; j++ {
				writeSource(t, filepath.Join(ws, "churn", "d"+string(rune('a'+j)), "m.ts"), "import y from 'zod'\n")
			}
			_ = os.RemoveAll(filepath.Join(ws, "churn"))
		}
	}()
	for i := 0; i < 200; i++ {
		if _, err := ScanImports(ws); err != nil {
			close(stop)
			wg.Wait()
			t.Fatalf("run %d: %v", i, err)
		}
	}
	close(stop)
	wg.Wait()
}
