//go:build e2e

package goldenpath

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// registries are throwaway npm and Python indexes the release publishes to.
type registries struct {
	npmURL, npmToken string
	pypiURL          string
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

func run(t *testing.T, dir string, argv ...string) string {
	t.Helper()
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", strings.Join(argv, " "), err, out)
	}
	return string(out)
}

func startServer(t *testing.T, dir string, argv ...string) {
	t.Helper()
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	log, _ := os.Create(filepath.Join(dir, "server.log"))
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		_ = log.Close()
		// A failed run shows what the server said: a registry that stopped
		// answering otherwise leaves only the client's ECONNREFUSED.
		if t.Failed() {
			if text, err := os.ReadFile(log.Name()); err == nil {
				t.Logf("%s server log:\n%s", filepath.Base(dir), text)
			}
		}
	})
}

func waitFor(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		if resp, err := http.Get(url); err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("%s did not come up", url)
}

func python() []string {
	if runtime.GOOS == "windows" {
		return []string{"py", "-3"}
	}
	return []string{"python3"}
}

// startRegistries installs and starts Verdaccio and pypiserver as single
// processes (no npx wrapper) so cleanup kills them on every OS.
func startRegistries(t *testing.T, base string) registries {
	t.Helper()
	npmDir := filepath.Join(base, "verdaccio")
	_ = os.MkdirAll(npmDir, 0o755)
	run(t, npmDir, "npm", "install", "--no-audit", "--no-fund", "--prefix", npmDir, "verdaccio@6")
	config := "storage: ./storage\nauth:\n  htpasswd:\n    file: ./htpasswd\n    max_users: 100\nuplinks: {}\npackages:\n  '**':\n    access: $all\n    publish: $authenticated\nlog: { type: stdout, format: pretty, level: warn }\n"
	_ = os.WriteFile(filepath.Join(npmDir, "config.yaml"), []byte(config), 0o644)
	npmPort := freePort(t)
	startServer(t, npmDir, "node", filepath.Join(npmDir, "node_modules", "verdaccio", "bin", "verdaccio"), "--config", "./config.yaml", "--listen", fmt.Sprintf("127.0.0.1:%d", npmPort))
	// 127.0.0.1, not localhost: Node may resolve localhost to ::1 first,
	// where nothing listens.
	npmURL := fmt.Sprintf("http://127.0.0.1:%d/", npmPort)
	waitFor(t, npmURL+"-/ping")
	req, _ := http.NewRequest(http.MethodPut, npmURL+"-/user/org.couchdb.user:ci", strings.NewReader(`{"name":"ci","password":"ci-pass"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var user struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&user)
	_ = resp.Body.Close()
	if user.Token == "" {
		t.Fatal("verdaccio did not issue a token")
	}

	pyDir := filepath.Join(base, "pypi")
	_ = os.MkdirAll(filepath.Join(pyDir, "packages"), 0o755)
	run(t, pyDir, append(python(), "-m", "venv", "venv")...)
	py := filepath.Join(pyDir, "venv", "bin", "python")
	if runtime.GOOS == "windows" {
		py = filepath.Join(pyDir, "venv", "Scripts", "python.exe")
	}
	run(t, pyDir, py, "-m", "pip", "install", "--quiet", "pypiserver", "passlib")
	pyPort := freePort(t)
	startServer(t, pyDir, py, "-m", "pypiserver", "run", "-i", "127.0.0.1", "-p", fmt.Sprint(pyPort), "-a", ".", "-P", ".", "packages")
	pypiURL := fmt.Sprintf("http://127.0.0.1:%d/", pyPort)
	waitFor(t, pypiURL)
	return registries{npmURL: npmURL, npmToken: user.Token, pypiURL: pypiURL}
}
