//go:build e2e

package goldenpath

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReleasePublishesADotnetTool releases a Go app as a .NET tool to a
// (fake) NuGet feed, then installs it from the pushed packages with
// `dotnet tool install` and runs the command it puts on PATH.
func TestReleasePublishesADotnetTool(t *testing.T) {
	if _, err := exec.LookPath("dotnet"); err != nil {
		t.Skip("the .NET SDK is not installed")
	}
	base := t.TempDir()
	feed := filepath.Join(base, "feed")
	if err := os.MkdirAll(feed, 0o755); err != nil {
		t.Fatal(err)
	}
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, header, err := r.FormFile("package")
		if r.Method != http.MethodPut || err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		keys = append(keys, r.Header.Get("X-NuGet-ApiKey"))
		raw, _ := io.ReadAll(f)
		if err := os.WriteFile(filepath.Join(feed, header.Filename), raw, 0o644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: append(identity(), "NUGET_API_KEY=key", "NUGET_PUSH_URL="+srv.URL)}
	w.smith(true, "new", "acme", "--ci", "github")
	w.dir = filepath.Join(base, "acme")
	w.smith(true, "add", "app", "tool", "--lang", "go")
	w.write("smith.toml", w.read("smith.toml")+
		"\n[projects.tool.release]\nplatforms = [\""+runtime.GOOS+"/"+runtime.GOARCH+"\"]\nnuget = \"acme-tool\"\ndescription = \"A tool\"\nlicense = \"MIT\"\n")
	remote := filepath.Join(base, "remote.git")
	run(t, base, "git", "init", "-q", "--bare", remote)
	w.git("remote", "add", "origin", remote)
	w.git("add", "-A")
	w.git("commit", "-qm", "feat: the tool")
	w.git("push", "-q", "origin", "main")

	out := w.smith(true, "release", "--yes")
	if !strings.Contains(out, "acme-tool.0.1.0.nupkg") {
		t.Fatalf("the release names no tool package:\n%s", out)
	}
	if len(keys) != 2 || keys[0] != "key" {
		t.Fatalf("pushed %d packages with keys %v:\n%s", len(keys), keys, out)
	}

	tools := filepath.Join(base, "tools")
	install := exec.Command("dotnet", "tool", "install", "acme-tool", "--add-source", feed, "--version", "0.1.0", "--tool-path", tools)
	install.Env = append(os.Environ(), "NUGET_PACKAGES="+filepath.Join(base, "nuget-packages"), "DOTNET_CLI_TELEMETRY_OPTOUT=1")
	if raw, err := install.CombinedOutput(); err != nil {
		t.Fatalf("dotnet tool install: %v\n%s", err, raw)
	}
	command := filepath.Join(tools, "tool")
	if runtime.GOOS == "windows" {
		command += ".cmd"
	}
	raw, err := exec.Command(command).CombinedOutput()
	if err != nil || !strings.Contains(string(raw), "world") {
		t.Fatalf("the installed tool: %v\n%s", err, raw)
	}
}
