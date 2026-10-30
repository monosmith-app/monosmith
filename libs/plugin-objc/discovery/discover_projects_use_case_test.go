package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const stringsPodspec = `# smithKind: package
# smithStack: apple
Pod::Spec.new do |s|
  s.name         = 'Strings'
  s.version      = "2.1.0"
  s.dependency 'Core/Base', '~> 1.0'
  s.dependency "AFNetworking"
  s.subspec 'Extras' do |ss|
    ss.dependency 'Other'
  end
end
`

func TestDiscover(t *testing.T) {
	ws := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		_ = os.WriteFile(p, []byte(body), 0o644)
	}
	write("libs/strings/Strings.podspec", stringsPodspec)
	write("libs/strings/Zeta.podspec", "Pod::Spec.new do |s|\n  s.name = 'Zeta'\nend\n")
	write("libs/core/Core.podspec", "Pod::Spec.new do |spec|\n  spec.name = 'Core'\n  spec.version = File.read('VERSION')\nend\n")
	write("libs/odd/Odd.podspec", "# nothing\n")
	write("Root.podspec", "")
	res, err := Discover(ws, []string{"Root.podspec", "libs/core/Core.podspec", "libs/odd/Odd.podspec", "libs/strings/Strings.podspec", "libs/strings/Zeta.podspec"})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	core, strs := res.Projects[0], res.Projects[1]
	if core.Name != "core" || core.Exports[0] != "Core" || core.Version != "" || core.Kind != protocol.ProjectLib || core.Tags[0] != NativeTag {
		t.Errorf("core %+v", core)
	}
	if strs.Name != "strings" || strs.Exports[0] != "Strings" || strs.Version != "2.1.0" || strs.Kind != protocol.ProjectPackage || strs.Stack != "apple" ||
		strings.Join(strs.Imports, ",") != "Core,AFNetworking,Other" {
		t.Errorf("the first podspec is the project's; dependencies reduce to pods: %+v", strs)
	}
	if got := strings.Join(strs.Tasks["lint"].Command, " "); got != "pod lib lint --quick Strings.podspec" {
		t.Errorf("lint %q", got)
	}
}

func TestSetVersionAndKind(t *testing.T) {
	out, ok := SetVersion(stringsPodspec, "2.2.0")
	if !ok || !strings.Contains(out, `s.version      = "2.2.0"`) {
		t.Errorf("the quotes and spacing are kept: %q", out)
	}
	if _, ok := SetVersion(out, "2.2.0"); ok {
		t.Error("the same version is no change")
	}
	if _, ok := SetVersion("s.version = File.read('VERSION')\n", "1.0.0"); ok {
		t.Error("a computed version is left alone")
	}
	if out, _ := SetKind(stringsPodspec, "lib"); !strings.HasPrefix(out, "# smithKind: lib\n# smithStack: apple\n") {
		t.Errorf("%q", out)
	}
	if out, _ := SetKind("Pod::Spec.new\r\n", "lib"); out != "# smithKind: lib\r\nPod::Spec.new\r\n" {
		t.Errorf("%q", out)
	}
}
