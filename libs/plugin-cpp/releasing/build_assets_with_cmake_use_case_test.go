package releasing

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// TestBuildAssetsWithRealCMake builds a tiny app with the real CMake, Ninja
// and compiler, unpacks the archive and runs the program in it. It is
// skipped where one of them is not installed.
func TestBuildAssetsWithRealCMake(t *testing.T) {
	compiler := false
	for _, c := range []string{"c++", "g++", "clang++", "cl"} {
		compiler = compiler || toolrun.Available(c)
	}
	if !toolrun.Available("cmake") || !toolrun.Available("ninja") || !compiler {
		t.Skip("cmake, ninja or a C++ compiler is not installed")
	}
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "hello")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.20)\nproject(acme_hello VERSION 3.4.5 LANGUAGES CXX)\nadd_executable(acme_hello main.cpp)\ntarget_compile_definitions(acme_hello PRIVATE VERSION=\"${PROJECT_VERSION}\")\n"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("#include <iostream>\nint main() { std::cout << \"hello \" << VERSION << std::endl; }\n"), 0o644)
	_ = os.WriteFile(filepath.Join(ws, "LICENSE"), []byte("MIT"), 0o644)

	project := protocol.Project{Name: "hello", Root: "apps/hello", Kind: protocol.ProjectApp, Tags: []string{discovery.NativeTag}}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	res, err := BuildAssets(ws, protocol.PublishParams{Project: project, Version: "3.4.5", Assets: spec(platform)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Binary != "acme_hello" || len(res.Assets) != 2 {
		t.Fatalf("%+v", res)
	}
	zr, err := zip.OpenReader(filepath.Join(ws, filepath.FromSlash(res.Assets[0])))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	out := t.TempDir()
	var program string
	for _, f := range zr.File {
		in, _ := f.Open()
		path := filepath.Join(out, f.Name)
		dst, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o755)
		_, _ = io.Copy(dst, in)
		_ = dst.Close()
		_ = in.Close()
		if strings.HasPrefix(f.Name, "acme_hello") {
			program = path
		}
	}
	if program == "" {
		t.Fatalf("no program in the archive of %d files", len(zr.File))
	}
	got, err := exec.Command(program).Output()
	if err != nil || strings.TrimSpace(string(got)) != "hello 3.4.5" {
		t.Fatalf("%q %v", got, err)
	}
}
