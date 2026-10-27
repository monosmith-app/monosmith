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

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestBuildAssetsWithRealDotnet publishes a tiny app for this machine's
// platform with the real SDK, unpacks the zip and runs the program in it,
// which must say the version stamped in. It is skipped where the SDK, or the
// runtime pack of this machine's platform (which a publish would download),
// is missing.
func TestBuildAssetsWithRealDotnet(t *testing.T) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	rid, err := RuntimeIdentifier(platform)
	if _, lookErr := exec.LookPath("dotnet"); lookErr != nil || err != nil {
		t.Skip("the .NET SDK is not installed")
	}
	if !RuntimePackCached(rid) {
		t.Skip("the " + rid + " runtime is not in the NuGet cache, and a test does not download it")
	}
	ws := t.TempDir()
	dir := filepath.Join(ws, "apps", "hello")
	_ = os.MkdirAll(dir, 0o755)
	version, _ := exec.Command("dotnet", "--version").Output()
	framework := "net" + strings.SplitN(strings.TrimSpace(string(version)), ".", 2)[0] + ".0"
	_ = os.WriteFile(filepath.Join(dir, "Hello.csproj"), []byte("<Project Sdk=\"Microsoft.NET.Sdk\"><PropertyGroup><OutputType>Exe</OutputType><TargetFramework>"+framework+"</TargetFramework></PropertyGroup></Project>"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "Program.cs"), []byte("System.Console.WriteLine(\"hello \" + typeof(Program).Assembly.GetName().Version!.ToString(3));\n"), 0o644)

	project := protocol.Project{Name: "hello", Root: "apps/hello", Kind: protocol.ProjectApp}
	res, err := BuildAssets(ws, protocol.PublishParams{Project: project, Version: "3.4.5", Assets: spec(platform)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Binary != "Hello" || len(res.Assets) != 2 {
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
		if strings.HasPrefix(f.Name, "Hello") {
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
