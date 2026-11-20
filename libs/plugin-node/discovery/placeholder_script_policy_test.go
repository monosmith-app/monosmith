package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPlaceholderScript(t *testing.T) {
	for _, command := range []string{
		`echo "Error: no test specified" && exit 1`,
		`echo 'Error: no test specified' && exit 1`,
		`echo \"Error: no test specified\" && exit 1`,
		`echo "Error: no test specified"; exit 1`,
		`echo "Error: no test specified" || exit 1`,
		`echo "Error: no test specified"`,
		`echo "Error: no test specified" && exit 0`,
		`  echo "error: No test specified"  &&  exit 1 `,
	} {
		if !IsPlaceholderScript(command) {
			t.Errorf("%q is the npm-init placeholder", command)
		}
	}
	for _, command := range []string{
		"jest",
		`echo "Error: no test specified" && exit 1 && jest`,
		`node test.js && echo "Error: no test specified"`,
		`echo done`,
	} {
		if IsPlaceholderScript(command) {
			t.Errorf("%q is a real script", command)
		}
	}
}

func TestDiscoverSkipsThePlaceholderTestScript(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "api")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"api","scripts":{"start":"node app.js","test":"echo \"Error: no test specified\" && exit 1"}}`), 0o644)

	res, err := Discover(ws, []string{"apps/api/package.json"})
	if err != nil {
		t.Fatal(err)
	}
	tasks := res.Projects[0].Tasks
	if _, ok := tasks["test"]; ok {
		t.Fatalf("the placeholder must give no test task: %+v", tasks["test"])
	}
	if _, ok := tasks["start"]; !ok {
		t.Fatal("a real script still gives a task")
	}
}
