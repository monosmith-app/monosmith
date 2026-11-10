//go:build e2e

package goldenpath

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// androidSDK is the Android SDK Gradle will use: ANDROID_HOME, else where
// Android Studio and the command-line tools install it.
func androidSDK() string {
	if dir := os.Getenv("ANDROID_HOME"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	candidates := map[string]string{
		"windows": filepath.Join(os.Getenv("LOCALAPPDATA"), "Android", "Sdk"),
		"darwin":  filepath.Join(home, "Library", "Android", "sdk"),
		"linux":   filepath.Join(home, "Android", "Sdk"),
	}
	if dir := candidates[runtime.GOOS]; dir != "" {
		if _, err := os.Stat(filepath.Join(dir, "platforms")); err == nil {
			return dir
		}
	}
	return ""
}

// TestAndroidWorkspace builds an Android app with the jvm plugin's android
// stack: AGP's build into dist/, unit tests, ktlint and Android Lint, the
// cache, doctor and CI.
func TestAndroidWorkspace(t *testing.T) {
	for _, tool := range []string{"java", "gradle"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " is not installed")
		}
	}
	sdk := androidSDK()
	if sdk == "" {
		t.Skip("the Android SDK is not installed")
	}
	base := t.TempDir()
	t.Cleanup(func() { _ = exec.Command("gradle", "--stop", "-q").Run() })
	w := &workspace{t: t, bin: buildSmith(t, base), dir: base, env: append(identity(), "ANDROID_HOME="+sdk)}
	w.smith(true, "new", "smith-e2e-android", "--ci", "github")
	w.dir = filepath.Join(base, "smith-e2e-android")
	w.smith(true, "add", "app", "mobile", "--stack", "android")
	if out := w.smith(false, "add", "lib", "ui", "--stack", "android"); !strings.Contains(out, "--stack kotlin") {
		t.Fatalf("an Android lib is refused with a hint:\n%s", out)
	}

	out := w.smith(true, "run", "lint", "test", "build")
	for _, want := range []string{"mobile:lint ✓", "mobile:test ✓", "mobile:build ✓"} {
		if !strings.Contains(out, want) {
			t.Fatalf("%s did not pass:\n%s", want, out)
		}
	}
	apk := filepath.Join(w.dir, "dist", "gradle", "smith-e2e-android-mobile", "outputs", "apk", "debug", "mobile-debug.apk")
	if _, err := os.Stat(apk); err != nil {
		t.Fatalf("the APK is built under dist/gradle: %v", err)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "dist", "apps", "mobile", "mobile-debug.apk")); err != nil {
		t.Errorf("the APK is shipped to dist/apps/mobile: %v", err)
	}
	if _, err := os.Stat(filepath.Join(w.dir, "apps", "mobile", "build")); err == nil {
		t.Error("apps/mobile/build exists: Gradle builds must go to dist/")
	}
	if out := w.smith(true, "run", "lint", "test"); !strings.Contains(out, "(3 from cache") {
		t.Fatalf("a re-run replays from the cache:\n%s", out)
	}
	if out := w.smith(true, "doctor"); !strings.Contains(out, "✓ the Android SDK is installed") {
		t.Fatalf("doctor:\n%s", out)
	}
	if ci := w.read(".github/workflows/ci.yml"); !strings.Contains(ci, "commandlinetools-linux-") {
		t.Fatalf("ci.yml installs the SDK where none is:\n%s", ci)
	}
}
