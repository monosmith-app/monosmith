package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func TestScanImportsSkipsSdistStagingFolders(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "src/pkg/a.py", "import requests\n")
	write(t, ws, "pkg-0.0.0/src/pkg/a.py", "import staged\n")
	got, err := ScanImports(ws)
	if err != nil || !reflect.DeepEqual(got, []string{"requests"}) {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestScanImportsIgnoresAFileThatVanishedBeforeItWasRead(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "a.py", "import requests\n")
	// A link to a file that is gone is listed by the walk and fails to open:
	// what the walk meets when a build deletes a file between listing and reading.
	if err := os.Symlink(filepath.Join(ws, "gone.py"), filepath.Join(ws, "b.py")); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	got, err := ScanImports(ws)
	if err != nil || !reflect.DeepEqual(got, []string{"requests"}) {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestScanImportsSurvivesFilesDeletedDuringTheWalk(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows reports a file being deleted as access denied, not as missing")
	}
	ws := t.TempDir()
	write(t, ws, "keep.py", "import requests\n")
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			for j := 0; j < 20; j++ {
				write(t, ws, filepath.Join("churn", "d"+string(rune('a'+j)), "m.py"), "import temp\n")
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
